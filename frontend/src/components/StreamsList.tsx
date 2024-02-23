import { gql } from "@/__generated__";
import { ApolloError, useQuery } from "@apollo/client";
import {
  TableBody,
  TableHead,
  Table,
  TableRow,
  TableCell,
  Box,
} from "@mui/material";
import Link from "next/link";
import { useState } from "react";
import ErrorAlert from "./ErrorAlert";

const STREAMS_QUERY = gql(/* GraphQL */ `
  query GetStreams($after: ID) {
    getStreams(after: $after) {
      edges {
        id
        createdAt
        displayName
        url
      }
      pageInfo {
        endCursor
      }
    }
  }
`);

type Stream = {
  id: string;
  createdAt: Date;
  displayName: string;
  url: string;
};

export default function StreamsList() {
  const [streams, setStreams] = useState<Stream[]>([]);
  const [errors, setErrors] = useState<ApolloError[]>([]);
  const { fetchMore } = useQuery(STREAMS_QUERY, {
    onError: e => {
      console.error(e);
      setErrors([...errors, e]);
    },
    onCompleted: data => {
      const {
        pageInfo: { endCursor },
        edges,
      } = data.getStreams;
      if (edges) {
        setStreams([...streams, ...edges]);
      }
      if (endCursor) {
        fetchMore({ variables: { after: endCursor } });
      }
    },
  });

  const rows = streams.map(s => (
    <TableRow key={s.id}>
      <TableCell>
        <Link href={`/streams/${s.id}`}>{s.displayName}</Link>
      </TableCell>
      <TableCell>
        <a target="_blank" href={s.url}>
          {s.url}
        </a>
      </TableCell>
      <TableCell>{new Date(s.createdAt).toLocaleString()}</TableCell>
    </TableRow>
  ));

  return (
    <>
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
            <TableCell>Name</TableCell>
            <TableCell>URL</TableCell>
            <TableCell>Created</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>{rows}</TableBody>
      </Table>
    </>
  );
}
