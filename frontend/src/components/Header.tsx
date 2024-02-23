"use client"

import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import AccountCircle from '@mui/icons-material/AccountCircle';
import LyricsIcon from '@mui/icons-material/Lyrics';
import { signIn, signOut, useSession } from 'next-auth/react';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import { useState } from 'react';
import Button from '@mui/material/Button';
import { Link } from '@mui/material';
import NextLink from "next/link";

export default function Header() {
  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static">
        <Toolbar>
          <Link href='/' component={NextLink} sx={{ color: "#fff", flexGrow: 1 }}>
            <IconButton
              size="large"
              edge="start"
              color="inherit"
              aria-label="menu"
              sx={{ mr: 2 }}
            >
              <LyricsIcon />
              <Typography variant="h6">
                Radio Chatter
              </Typography>
            </IconButton>
          </Link>

          <CurrentUserBadge />
        </Toolbar>
      </AppBar>
    </Box>
  );
}

function CurrentUserBadge() {
  const session = useSession();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);

  const handleMenu = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  switch (session.status) {
    case "authenticated":
      return (
        <div>
          <IconButton
            size="large"
            aria-label="account of current user"
            aria-controls="menu-appbar"
            aria-haspopup="true"
            onClick={handleMenu}
            color="inherit"
          >
            <AccountCircle />
            <Typography>
              {session.data.user?.name}
            </Typography>
          </IconButton>
          <Menu
            id="menu-appbar"
            anchorEl={anchorEl}
            anchorOrigin={{
              vertical: 'top',
              horizontal: 'right',
            }}
            keepMounted
            transformOrigin={{
              vertical: 'top',
              horizontal: 'right',
            }}
            open={Boolean(anchorEl)}
            onClose={handleClose}
          >
            <MenuItem onClick={() => signOut()}>Log Out</MenuItem>
          </Menu>
        </div>
      )

    case "loading":
    case "unauthenticated":
      return (
        <div>
          <Button sx={{ color: "#fff" }} onClick={() => signIn()}>
            Login
          </Button>
        </div>
      )

  }
}
