# Radio Chatter

[![Continuous Integration](https://github.com/Michael-F-Bryan/radio-chatter/actions/workflows/ci.yml/badge.svg)](https://github.com/Michael-F-Bryan/radio-chatter/actions/workflows/ci.yml)
[![Link to the production environment](https://img.shields.io/badge/Frontend-live-green)](https://radio-chatter.vercel.app/)

Radio Chatter is a service created by the [Communications Support Unit][csu] for
monitoring DFES radio traffic and providing better situational awareness during
emergencies.

## Architecture

This project contains 4 major components,

- Download - downloads an audio stream using `ffmpeg` and splits it into both
  60-second chunks and "transmissions" containing actual speech
- Transcribe - takes a transmission and runs speech-to-text on it
- Backend - A GraphQL API that gives users access to everything (see
  [`schema.graphql`](pkg/graphql/schema.graphql))
- Frontend - A NextJS UI that people can use to listen to streams and search
  through transmissions as they are recorded and transcribed

## License

Radio Chatter is proprietary software developed for and by the Communications
Support Unit (CSU). At this time, usage and access to the project source code
are strictly limited to authorized CSU members and affiliates. Any
redistribution, modification, or use of the software outside of CSU is
prohibited without explicit permission from the project administrators.

For inquiries regarding licensing or permissions, please contact the project
team directly.

[csu]: https://csu-ses.com.au/
