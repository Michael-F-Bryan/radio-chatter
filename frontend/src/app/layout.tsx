import type { Metadata } from "next";
import { Inter } from "next/font/google";
import { loadErrorMessages, loadDevMessages } from "@apollo/client/dev";

import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";

import "./globals.css";
import Header from "@/components/Header";
import Providers from "../components/Providers";
import RequiresAuth from "@/components/RequiresAuth";

// Make Apollo include useful error messages by default
loadDevMessages();
loadErrorMessages();

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Radio Chatter",
  description:
    "A service created by the Communications Support Unit for monitoring DFES radio traffic and providing better situational awareness during emergencies.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <Providers>
          <Header />
          <RequiresAuth sx={{ mx: "auto", mt: "2em" }}>{children}</RequiresAuth>
        </Providers>
      </body>
    </html>
  );
}
