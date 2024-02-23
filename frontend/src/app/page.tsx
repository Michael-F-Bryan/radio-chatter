"use client";

import Container from "@mui/material/Container";
import { Typography } from "@mui/material";
import StreamsList from "@/components/StreamsList";

export default function Home() {
  return (
    <Container component="main">
      <Typography variant="h1">Streams</Typography>
      <StreamsList />
    </Container>
  )
}
