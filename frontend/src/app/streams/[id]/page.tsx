"use client";
import { useSubscription } from "@apollo/client";

import { gql } from "@/__generated__";
import { useState } from "react";
import { Container, Table, TableBody, TableCell, TableHead, TableRow } from "@mui/material";

const TRANSMISSION_SUBSCRIPTION = gql(/* GraphQL */ `
  subscription transmissions {
    transmission {
      id
      timestamp
      downloadUrl
      transcription {
        content
      }
    }
  }
`);



type Record = {
  id: string;
  timestamp: Date;
  downloadUrl?: string;
  content?: string;
};

export default function Stream() {
  const [transmissions, setTransmissions] = useState<Record[]>([]);

  const { loading, data, error, variables } = useSubscription(
    TRANSMISSION_SUBSCRIPTION,
    {
      onData: opts => {
        const t = opts.data.data?.transmission;
        if (t) {
          const value: Record = {
            id: t.id,
            timestamp: t.timestamp,
            downloadUrl: t.downloadUrl || undefined,
            content: t.transcription?.content,
          };
          setTransmissions([...transmissions, value]);
        }
      },
    },
  );

  console.log(loading, data, error, variables);

  const rows = transmissions.map(t => (
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
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>ID</TableCell>
            <TableCell>Message</TableCell>
            <TableCell>Timestamp</TableCell>
            <TableCell></TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {rows}
        </TableBody>
      </Table>
    </Container>
  );
}
