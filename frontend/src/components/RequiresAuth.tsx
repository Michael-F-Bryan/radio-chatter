"use client";

import {
  Button,
  CircularProgress,
  Container,
  ContainerTypeMap,
} from "@mui/material";
import { DefaultComponentProps } from "@mui/material/OverridableComponent";
import { signIn, useSession } from "next-auth/react";
import { ReactNode, useEffect } from "react";

export default function RequiresAuth({
  children,
  ...props
}: React.PropsWithChildren<DefaultComponentProps<ContainerTypeMap>>) {
  const { status } = useSession();

  useEffect(() => {
    if (status == "unauthenticated") {
      signIn();
    }
  }, [status]);

  let content: ReactNode;

  switch (status) {
    case "authenticated":
      content = children;
      break;
    case "loading":
      content = <CircularProgress variant="indeterminate" />;
      break;
    case "unauthenticated":
      content = <Button onClick={() => signIn()}>Please Sign In.</Button>;
      break;
  }

  return <Container {...props}>{content}</Container>;
}
