"use client";
import { ApolloError, useSubscription } from "@apollo/client";

import { gql } from "@/__generated__";
import { useMemo, useState } from "react";
import {
  Box,
  Container,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from "@mui/material";
import ErrorAlert from "./ErrorAlert";

const TRANSMISSION_SUBSCRIPTION = gql(/* GraphQL */ `
  subscription transmissions {
    transmission {
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
    transcription {
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

type Message = {
  id: string;
  timestamp: Date;
  downloadUrl?: string;
  content?: string;
};

type Props = {
  streamID: string;
};

/**
 * A table which will automatically display new transmissions and transcriptions
 * for a particular stream as they are generated.
 */
export default function StreamTable({ streamID, }: Props) {
  // TODO: populate the state with the last ~6 hours worth of transmissions
  const [transmissions, setTransmissions] = useState<Record<string, Message>>(
    {},
  );
  const sortedTransmissions = useMemo(() => {
    // FIXME: Figure out how to store this in a sorted way.
    return Object.values(transmissions).toSorted(
      (a, b) => a.timestamp.valueOf() - b.timestamp.valueOf(),
    );
  }, [transmissions]);
  const [errors, setErrors] = useState<ApolloError[]>([]);

  // Capture new transmissions as they are recorded
  useSubscription(TRANSMISSION_SUBSCRIPTION, {
    onData: opts => {
      const t = opts.data?.data?.transmission;
      if (t && t.chunk.stream.id == streamID) {
        setTransmissions({
          ...transmissions,
          [t.id]: {
            id: t.id,
            timestamp: new Date(t.timestamp),
            downloadUrl: t.downloadUrl || undefined,
            content: t.transcription?.content,
          },
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
        setTransmissions({
          ...transmissions,
          [t.id]: {
            id: t.id,
            timestamp: new Date(t.transmission.timestamp),
            downloadUrl: t.transmission.downloadUrl || undefined,
            content: t.content,
          },
        });
      }
    },
    onError: e => {
      console.error(e);
      setErrors([...errors, e]);
    },
  });

  const rows = sortedTransmissions.map(t => (
    <TableRow key={t.id}>
      <TableCell>{t.id}</TableCell>
      <TableCell>{t.content || "-"}</TableCell>
      <TableCell>{t.timestamp.toString()}</TableCell>
      <TableCell>
        <a href={t.downloadUrl} target="_blank">
          Link
        </a>
      </TableCell>
    </TableRow>
  ));

  return (
    <Container component="main">
      <Box>
        {errors.map((e, i) => (
          <ErrorAlert
            key={i}
            error={e}
            dismiss={() => setErrors(errors.filter((_, index) => index != i))}
          />
        ))}
      </Box>

      <Table>
        <TableHead>
          <TableRow>
            <TableCell>ID</TableCell>
            <TableCell>Message</TableCell>
            <TableCell>Timestamp</TableCell>
            <TableCell></TableCell>
          </TableRow>
        </TableHead>
        <TableBody>{rows}</TableBody>
      </Table>
    </Container>
  );
}
