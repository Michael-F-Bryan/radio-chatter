"use client";
import { useSubscription } from "@apollo/client";

import styles from "./page.module.css";
import { gql } from "../__generated__/gql";
import { useState } from "react";

const TRANSMISSION_SUBSCRIPTION = gql(/* GraphQL */ `
  subscription transmissions {
    transmission {
      id
      timestamp
      downloadUrl
    }
  }
`);
type Record = {
  id: string,
  timestamp: Date,
  downloadUrl?: string,
}

export default function Home() {
  const [transmissions, setTransmissions] = useState<Record[]>([]);

  const { loading, data, error, variables } = useSubscription(
    TRANSMISSION_SUBSCRIPTION,
    {
      onData: (opts) => {
        const t = opts.data.data?.transmission;
        if (t) {
          const value: Record = {
            id: t.id, timestamp: t.timestamp, downloadUrl: t.downloadUrl || undefined,
          }
          setTransmissions([...transmissions, value]);
        }

      }
    }
  );

  console.log(loading, data, error, variables);

  const rows = transmissions.map(t => (
    <tr key={t.id}>
      <td>{t.id}</td>
      <td>{t.timestamp.toString()}</td>
      <td><a href={t.downloadUrl} target="_blank">Link</a></td>
    </tr>
  ));

  return (
    <main className={styles.main}>
      <h1>Main page!</h1>
      <table>
        <thead>
          <tr>
            <th>#</th>
            <th>Timestamp</th>
            <th>Link</th>
          </tr>
        </thead>
        <tbody>
          {rows}
        </tbody>
      </table>
    </main>
  );
}
