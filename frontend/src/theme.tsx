import { createTheme } from "@mui/material/styles";
import NextLink from "next/link";
import { forwardRef } from "react";

interface LinkProps {
  href: string;
  [name: string]: unknown;
}

export const LinkBehaviour = forwardRef<HTMLAnchorElement, LinkProps>(
  (props, ref) => {
    return <NextLink ref={ref} {...props} />;
  },
);

export const theme = createTheme({
  components: {
    MuiLink: {
      defaultProps: {
        component: LinkBehaviour,
      },
    },
    MuiButtonBase: {
      defaultProps: {
        LinkComponent: LinkBehaviour,
      },
    },
  },
});
