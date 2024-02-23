import { gql } from "@/__generated__";
import { useLazyQuery } from "@apollo/client";
import { TableBody, TableHead, Table, TableRow, TableCell } from "@mui/material";
import Link from "next/link";
import { useEffect } from "react"

const STREAMS_QUERY = gql(/* GraphQL */ `
  query GetStreams($after: ID) {
      getStreams(after: $after) {
        edges {
          id
          createdAt
          displayName
          url
        }
      }
    }
`);


export default function StreamsList() {
    const [execute, { loading, data, error, called, fetchMore, refetch }] = useLazyQuery(STREAMS_QUERY);

    useEffect(() => {
        if (!called && !loading) {
            // Make sure we automatically run the query when the page is loaded
            console.log("Triggering query")
            execute();
        }
    }, [called, loading]);

    const streams = data?.getStreams?.edges;

    useEffect(() => {
        // Note: automatically keep fetching more streams until nothing changes
        const last = streams?.at(-1)?.id;
        if (called && last) {
            console.log("Fetching more", last);
            fetchMore({ variables: { after: last } });
        }
    }, [streams]);


    const rows = (streams || []).map(s => (
        <TableRow key={s.id}>
            <TableCell><Link href={`/streams/${s.id}`}>{s.displayName}</Link></TableCell>
            <TableCell><a target="_blank" href={s.url}>{s.url}</a></TableCell>
            <TableCell>{new Date(s.createdAt).toLocaleString()}</TableCell>
        </TableRow>
    ));

    return (
        <Table>
            <TableHead>
                <TableRow>
                    <TableCell>Name</TableCell>
                    <TableCell>URL</TableCell>
                    <TableCell>Created</TableCell>
                </TableRow>
            </TableHead>
            <TableBody>
                {rows}
            </TableBody>
        </Table>
    )

}
