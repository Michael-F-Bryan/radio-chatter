"use client";
import {
  ApolloError,
  useLazyQuery,
  useSubscription,
} from "@apollo/client";
import dayjs, { Dayjs } from "dayjs";
import { useEffect, useMemo, useState } from "react";
import {
  Box,
  ButtonGroup,
  Container,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from "@mui/material";
import { Download, PlayArrow } from "@mui/icons-material";

import { gql } from "@/__generated__";
import { ExistingTransmissionsQueryVariables } from "@/__generated__/graphql";
import ErrorAlert from "./ErrorAlert";

const TRANSMISSION_SUBSCRIPTION = gql(/* GraphQL */ `
  subscription transmissions {
    allTransmissions {
      id
      timestamp
      downloadUrl
      transcription {
        content
      }
      chunk {
        stream {
          id
        }
      }
    }
  }
`);

const TRANSCRIPTION_SUBSCRIPTION = gql(/* GraphQL */ `
  subscription transcriptions {
    allTranscriptions {
      id
      content
      transmission {
        id
        timestamp
        downloadUrl
        chunk {
          stream {
            id
          }
        }
      }
    }
  }
`);

const EXISTING_TRANSMISSIONS_QUERY = gql(/* GraphQL */ `
  query existingTransmissions($stream: ID!, $createdAfter: Time, $after: ID) {
    getStreamById(id: $stream) {
      id
      transmissions(createdAfter: $createdAfter, after: $after) {
        edges {
          id
          timestamp
          downloadUrl
          transcription {
            content
          }
        }
        pageInfo {
          endCursor
        }
      }
    }
  }
`);

type Message = {
  id: string;
  timestamp: Dayjs;
  downloadUrl?: string;
  content?: string;
};

type Props = {
  streamID: string;
};

const CONTEXT_THRESHOLD_MS = 6 * 3600 * 1000;

/**
 * A table which will automatically display new transmissions and transcriptions
 * for a particular stream as they are generated.
 */
export default function StreamTable({ streamID }: Props) {
  const [variables, setVariables] = useState<
    ExistingTransmissionsQueryVariables | undefined
  >(() => ({
    stream: streamID,
    createdAfter: new Date(Date.now() - CONTEXT_THRESHOLD_MS),
  }));
  const [messages, setMessages] = useState<Record<string, Message>>({});
  const sorted = useMemo(() => {
    // FIXME: Figure out how to store this in a sorted way.
    return Object.values(messages).toSorted(
      (a, b) => b.timestamp.valueOf() - a.timestamp.valueOf(),
    );
  }, [messages]);
  const [errors, setErrors] = useState<ApolloError[]>([]);

  // First, we load some historic data to provide a bit of context
  const [execute, { loading, data, error }] = useLazyQuery(
    EXISTING_TRANSMISSIONS_QUERY,
  );

  useEffect(() => {
    if (!loading && !error && variables) {
      execute({ variables })
        .then(({ data }) => {
          if (!data?.getStreamById) {
            return;
          }

          const {
            edges,
            pageInfo: { endCursor },
          } = data.getStreamById.transmissions;

          if (edges) {
            setMessages(extendMessages(messages, edges));
          }

          if (endCursor) {
            const variables = { stream: streamID, after: endCursor };
            console.log("[Updating variables]", variables);
            setVariables(variables);
          } else {
            console.log("[Done]");
            setVariables(undefined);
          }
        })
        .catch(e => {
          console.error(e);
          setErrors([...errors, e]);
        });
    }
  }, [variables, loading, error]);

  if (error && !errors.includes(error)) {
    setErrors([...errors, error]);
  }

  console.log("[Data]", {
    loading,
    edges: data?.getStreamById?.transmissions?.edges,
  });

  // Capture new transmissions as they are recorded
  useSubscription(TRANSMISSION_SUBSCRIPTION, {
    onData: opts => {
      const t = opts.data?.data?.transmission;
      if (t && t.chunk.stream.id == streamID) {
        const msg = {
          id: t.id,
          timestamp: dayjs(t.timestamp),
          downloadUrl: t.downloadUrl || undefined,
          content: t.transcription?.content,
        };
        console.log("[Detected]", msg);
        setMessages({
          ...messages,
          [t.id]: msg,
        });
      }
    },
    onError: e => {
      console.error(e);
      setErrors([...errors, e]);
    },
  });

  // Make sure we update a row whenever speech-to-text runs on it
  useSubscription(TRANSCRIPTION_SUBSCRIPTION, {
    onData: opts => {
      const t = opts.data?.data?.transcription;
      if (t && t.transmission.chunk.stream.id == streamID) {
        const msg = {
          id: t.id,
          timestamp: dayjs(t.transmission.timestamp),
          downloadUrl: t.transmission.downloadUrl || undefined,
          content: t.content,
        };
        console.log("[Speech to text]", msg);
        setMessages({
          ...messages,
          [t.id]: msg,
        });
      }
    },
    onError: e => {
      console.error(e);
      setErrors([...errors, e]);
    },
  });

  const play = (msg: Message) => {
    console.log("[Play]", msg);
  };

  const rows = sorted.map(t => (
    <TableRow key={t.id}>
      <TableCell title={t.timestamp.toString()}>
        {t.timestamp.format("HH:mm:ss")}
      </TableCell>
      <TableCell>{t.content || "-"}</TableCell>
      <TableCell>
        <ButtonGroup>
          <IconButton onClick={() => play(t)}>
            <PlayArrow />
          </IconButton>
          <IconButton component="a" href={t.downloadUrl} target="_blank">
            <Download sx={{ my: "auto" }} />
          </IconButton>
        </ButtonGroup>
      </TableCell>
    </TableRow>
  ));

  return (
    <Container component="main" sx={{ width: "100%" }}>
      <Box>
        {errors.map((e, i) => (
          <ErrorAlert
            key={i}
            error={e}
            dismiss={() => setErrors(errors.filter((_, index) => index != i))}
          />
        ))}
      </Box>

      <Table size="small" stickyHeader>
        <TableHead>
          <TableRow>
            <TableCell>Timestamp</TableCell>
            <TableCell>Message</TableCell>
            <TableCell></TableCell>
          </TableRow>
        </TableHead>
        <TableBody>{rows}</TableBody>
      </Table>
    </Container>
  );
}

type Edge = {
  id: string;
  timestamp: any;
  transcription?: { content: string } | null;
  downloadUrl?: string | null | undefined;
};

function extendMessages(
  transmissions: Readonly<Record<string, Readonly<Message>>>,
  edges: Edge[],
) {
  const newTransmissions: Record<string, Message> = Object.fromEntries(
    edges.map(edge => {
      const msg: Message = {
        id: edge.id,
        timestamp: dayjs(edge.timestamp),
        content: edge.transcription?.content,
        downloadUrl: edge.downloadUrl || undefined,
      };
      return [edge.id, msg];
    }),
  );

  return { ...transmissions, ...newTransmissions };
}
