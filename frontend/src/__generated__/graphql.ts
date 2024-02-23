/* eslint-disable */
import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  /** A RFC3339 timestamp (e.g. "2006-01-02T15:04:05.999999999Z07:00"). */
  Time: { input: any; output: any; }
};

export type Chunk = Node & {
  __typename?: 'Chunk';
  createdAt: Scalars['Time']['output'];
  /** Where the chunk's audio file can be downloaded from. */
  downloadUrl?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  /** A SHA-256 checksum of the chunk's audio file. */
  sha256: Scalars['String']['output'];
  /** The stream this chunk belongs to. */
  stream: Stream;
  /** When the chunk was first broadcast. */
  timestamp: Scalars['Time']['output'];
  /** Iterate over the radio messages detected in the chunk. */
  transmissions: TransmissionsConnection;
  updatedAt: Scalars['Time']['output'];
};


export type ChunkTransmissionsArgs = {
  after?: InputMaybe<Scalars['ID']['input']>;
  count?: Scalars['Int']['input'];
  createdAfter?: InputMaybe<Scalars['Time']['input']>;
};

export type ChunksConnection = {
  __typename?: 'ChunksConnection';
  edges?: Maybe<Array<Chunk>>;
  pageInfo: PageInfo;
};

export type Mutation = {
  __typename?: 'Mutation';
  /** Register a new stream. */
  registerStream: Stream;
  /** Remove a stream. */
  removeStream: Stream;
};


export type MutationRegisterStreamArgs = {
  input: RegisterStreamVariables;
};


export type MutationRemoveStreamArgs = {
  id: Scalars['ID']['input'];
};

/** Properties shared by all items stored in the database. */
export type Node = {
  /** When the item was created. */
  createdAt: Scalars['Time']['output'];
  /** A unique ID for this item. */
  id: Scalars['ID']['output'];
  /** When the item was last updated. */
  updatedAt: Scalars['Time']['output'];
};

/** Information about a page in a paginated query. */
export type PageInfo = {
  __typename?: 'PageInfo';
  /** A cursor that can be passed as an "after" parameter to read the next page. */
  endCursor?: Maybe<Scalars['ID']['output']>;
  /** Are there any more pages? */
  hasNextPage: Scalars['Boolean']['output'];
  /** How many items were in this page? */
  length: Scalars['Int']['output'];
};

export type Query = {
  __typename?: 'Query';
  /** Look up a chunk by its ID. */
  getChunkById?: Maybe<Chunk>;
  /** Look up a stream by its ID. */
  getStreamById?: Maybe<Stream>;
  /** Iterate over all active streams. */
  getStreams: StreamsConnection;
  /** Look up a transmission by its ID. */
  getTransmissionById?: Maybe<Transmission>;
};


export type QueryGetChunkByIdArgs = {
  id: Scalars['ID']['input'];
};


export type QueryGetStreamByIdArgs = {
  id: Scalars['ID']['input'];
};


export type QueryGetStreamsArgs = {
  after?: InputMaybe<Scalars['ID']['input']>;
  count?: Scalars['Int']['input'];
  createdAfter?: InputMaybe<Scalars['Time']['input']>;
};


export type QueryGetTransmissionByIdArgs = {
  id: Scalars['ID']['input'];
};

export type RegisterStreamVariables = {
  displayName: Scalars['String']['input'];
  url: Scalars['String']['input'];
};

/** A stream to monitor and extract transmissions from. */
export type Stream = Node & {
  __typename?: 'Stream';
  /** Iterate over the raw chunks of audio downloaded for this stream. */
  chunks: ChunksConnection;
  createdAt: Scalars['Time']['output'];
  /** The human-friendly name for this stream. */
  displayName: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  /** Iterate over the radio messages detected in the stream. */
  transmissions: TransmissionsConnection;
  updatedAt: Scalars['Time']['output'];
  /**
   * Where the stream can be downloaded from.
   *
   * This is typically a URL, but can technically be anything ffmpeg allows as an
   * input.
   */
  url: Scalars['String']['output'];
};


/** A stream to monitor and extract transmissions from. */
export type StreamChunksArgs = {
  after?: InputMaybe<Scalars['ID']['input']>;
  count?: Scalars['Int']['input'];
  createdAfter?: InputMaybe<Scalars['Time']['input']>;
};


/** A stream to monitor and extract transmissions from. */
export type StreamTransmissionsArgs = {
  after?: InputMaybe<Scalars['ID']['input']>;
  count?: Scalars['Int']['input'];
  createdAfter?: InputMaybe<Scalars['Time']['input']>;
};

export type StreamsConnection = {
  __typename?: 'StreamsConnection';
  edges?: Maybe<Array<Stream>>;
  pageInfo: PageInfo;
};

export type Subscription = {
  __typename?: 'Subscription';
  /** Get chunks as they are recorded. */
  chunk: Chunk;
  /** Get transcriptions as transmissions are processed with speech-to-text. */
  transcription: Transcription;
  /** Get transmissions as they are detected. */
  transmission: Transmission;
};

export type Transcription = Node & {
  __typename?: 'Transcription';
  content: Scalars['String']['output'];
  createdAt: Scalars['Time']['output'];
  id: Scalars['ID']['output'];
  /** The transmission this transcription belongs to. */
  transmission: Transmission;
  updatedAt: Scalars['Time']['output'];
};

/** A radio transmission. */
export type Transmission = Node & {
  __typename?: 'Transmission';
  /** The chunk this transmission belongs to. */
  chunk: Chunk;
  createdAt: Scalars['Time']['output'];
  /** Where the chunk's audio file can be downloaded from. */
  downloadUrl?: Maybe<Scalars['String']['output']>;
  id: Scalars['ID']['output'];
  /** How long is the transmission, in seconds? */
  length: Scalars['Float']['output'];
  /** A SHA-256 checksum of the chunk's audio file. */
  sha256: Scalars['String']['output'];
  /** When the transmission was first broadcast. */
  timestamp: Scalars['Time']['output'];
  transcription?: Maybe<Transcription>;
  updatedAt: Scalars['Time']['output'];
};

export type TransmissionsConnection = {
  __typename?: 'TransmissionsConnection';
  edges?: Maybe<Array<Transmission>>;
  pageInfo: PageInfo;
};

export type StreamQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type StreamQuery = { __typename?: 'Query', getStreamById?: { __typename?: 'Stream', id: string, displayName: string } | null };

export type TransmissionsSubscriptionVariables = Exact<{ [key: string]: never; }>;


export type TransmissionsSubscription = { __typename?: 'Subscription', transmission: { __typename?: 'Transmission', id: string, timestamp: any, downloadUrl?: string | null, transcription?: { __typename?: 'Transcription', content: string } | null, chunk: { __typename?: 'Chunk', stream: { __typename?: 'Stream', id: string } } } };

export type TranscriptionsSubscriptionVariables = Exact<{ [key: string]: never; }>;


export type TranscriptionsSubscription = { __typename?: 'Subscription', transcription: { __typename?: 'Transcription', id: string, content: string, transmission: { __typename?: 'Transmission', id: string, timestamp: any, downloadUrl?: string | null, chunk: { __typename?: 'Chunk', stream: { __typename?: 'Stream', id: string } } } } };

export type ExistingTransmissionsQueryVariables = Exact<{
  stream: Scalars['ID']['input'];
  createdAfter?: InputMaybe<Scalars['Time']['input']>;
  after?: InputMaybe<Scalars['ID']['input']>;
}>;


export type ExistingTransmissionsQuery = { __typename?: 'Query', getStreamById?: { __typename?: 'Stream', id: string, transmissions: { __typename?: 'TransmissionsConnection', edges?: Array<{ __typename?: 'Transmission', id: string, timestamp: any, downloadUrl?: string | null, transcription?: { __typename?: 'Transcription', content: string } | null }> | null, pageInfo: { __typename?: 'PageInfo', endCursor?: string | null } } } | null };

export type GetStreamsQueryVariables = Exact<{
  after?: InputMaybe<Scalars['ID']['input']>;
}>;


export type GetStreamsQuery = { __typename?: 'Query', getStreams: { __typename?: 'StreamsConnection', edges?: Array<{ __typename?: 'Stream', id: string, createdAt: any, displayName: string, url: string }> | null, pageInfo: { __typename?: 'PageInfo', endCursor?: string | null } } };


export const StreamDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"stream"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"id"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"getStreamById"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"id"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"displayName"}}]}}]}}]} as unknown as DocumentNode<StreamQuery, StreamQueryVariables>;
export const TransmissionsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"transmissions"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"transmission"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"downloadUrl"}},{"kind":"Field","name":{"kind":"Name","value":"transcription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"content"}}]}},{"kind":"Field","name":{"kind":"Name","value":"chunk"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"stream"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}}]}}]}}]}}]}}]} as unknown as DocumentNode<TransmissionsSubscription, TransmissionsSubscriptionVariables>;
export const TranscriptionsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"subscription","name":{"kind":"Name","value":"transcriptions"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"transcription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"content"}},{"kind":"Field","name":{"kind":"Name","value":"transmission"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"downloadUrl"}},{"kind":"Field","name":{"kind":"Name","value":"chunk"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"stream"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}}]}}]}}]}}]}}]}}]} as unknown as DocumentNode<TranscriptionsSubscription, TranscriptionsSubscriptionVariables>;
export const ExistingTransmissionsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"existingTransmissions"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"stream"}},"type":{"kind":"NonNullType","type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"createdAfter"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"Time"}}},{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"after"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"getStreamById"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"id"},"value":{"kind":"Variable","name":{"kind":"Name","value":"stream"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"transmissions"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"createdAfter"},"value":{"kind":"Variable","name":{"kind":"Name","value":"createdAfter"}}},{"kind":"Argument","name":{"kind":"Name","value":"after"},"value":{"kind":"Variable","name":{"kind":"Name","value":"after"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"edges"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"timestamp"}},{"kind":"Field","name":{"kind":"Name","value":"downloadUrl"}},{"kind":"Field","name":{"kind":"Name","value":"transcription"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"content"}}]}}]}},{"kind":"Field","name":{"kind":"Name","value":"pageInfo"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"endCursor"}}]}}]}}]}}]}}]} as unknown as DocumentNode<ExistingTransmissionsQuery, ExistingTransmissionsQueryVariables>;
export const GetStreamsDocument = {"kind":"Document","definitions":[{"kind":"OperationDefinition","operation":"query","name":{"kind":"Name","value":"GetStreams"},"variableDefinitions":[{"kind":"VariableDefinition","variable":{"kind":"Variable","name":{"kind":"Name","value":"after"}},"type":{"kind":"NamedType","name":{"kind":"Name","value":"ID"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"getStreams"},"arguments":[{"kind":"Argument","name":{"kind":"Name","value":"after"},"value":{"kind":"Variable","name":{"kind":"Name","value":"after"}}}],"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"edges"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"id"}},{"kind":"Field","name":{"kind":"Name","value":"createdAt"}},{"kind":"Field","name":{"kind":"Name","value":"displayName"}},{"kind":"Field","name":{"kind":"Name","value":"url"}}]}},{"kind":"Field","name":{"kind":"Name","value":"pageInfo"},"selectionSet":{"kind":"SelectionSet","selections":[{"kind":"Field","name":{"kind":"Name","value":"endCursor"}}]}}]}}]}}]} as unknown as DocumentNode<GetStreamsQuery, GetStreamsQueryVariables>;