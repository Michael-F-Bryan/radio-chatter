import NextAuth, { User } from "next-auth";
import CredentialsProvider from "next-auth/providers/credentials";

const handler = NextAuth({
  providers: [
    // TODO: swap this out for a proper auth provider
    CredentialsProvider({
      name: "Credentials",
      credentials: {
        username: { label: "Username", type: "text", placeholder: "jsmith" },
        password: { label: "Password", type: "password" },
      },
      authorize: async (credentials, req): Promise<User | null> => {
        const username = credentials?.username;

        if (username) {
          return { id: username, name: username };
        } else {
          return null;
        }
      },
    }),
  ],
});

export { handler as GET, handler as POST };
