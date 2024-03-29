/* eslint-disable */
import * as types from './graphql';
import { TypedDocumentNode as DocumentNode } from '@graphql-typed-document-node/core';

/**
 * Map of all GraphQL operations in the project.
 *
 * This map has several performance disadvantages:
 * 1. It is not tree-shakeable, so it will include all operations in the project.
 * 2. It is not minifiable, so the string of a GraphQL query will be multiple times inside the bundle.
 * 3. It does not support dead code elimination, so it will add unused operations.
 *
 * Therefore it is highly recommended to use the babel or swc plugin for production.
 */
const documents = {
    "\n  query stream($id: ID!) {\n    getStreamById(id: $id) {\n      id\n      displayName\n    }\n  }\n": types.StreamDocument,
    "\n  subscription transmissions {\n    allTransmissions {\n      id\n      timestamp\n      downloadUrl\n      transcription {\n        content\n      }\n      chunk {\n        stream {\n          id\n        }\n      }\n    }\n  }\n": types.TransmissionsDocument,
    "\n  subscription transcriptions {\n    allTranscriptions {\n      id\n      content\n      transmission {\n        id\n        timestamp\n        downloadUrl\n        chunk {\n          stream {\n            id\n          }\n        }\n      }\n    }\n  }\n": types.TranscriptionsDocument,
    "\n  query existingTransmissions($stream: ID!, $createdAfter: Time, $after: ID) {\n    getStreamById(id: $stream) {\n      id\n      transmissions(createdAfter: $createdAfter, after: $after) {\n        edges {\n          id\n          timestamp\n          downloadUrl\n          transcription {\n            content\n          }\n        }\n        pageInfo {\n          endCursor\n        }\n      }\n    }\n  }\n": types.ExistingTransmissionsDocument,
    "\n  query GetStreams($after: ID) {\n    getStreams(after: $after) {\n      edges {\n        id\n        createdAt\n        displayName\n        url\n      }\n      pageInfo {\n        endCursor\n      }\n    }\n  }\n": types.GetStreamsDocument,
};

/**
 * The gql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 *
 *
 * @example
 * ```ts
 * const query = gql(`query GetUser($id: ID!) { user(id: $id) { name } }`);
 * ```
 *
 * The query argument is unknown!
 * Please regenerate the types.
 */
export function gql(source: string): unknown;

/**
 * The gql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function gql(source: "\n  query stream($id: ID!) {\n    getStreamById(id: $id) {\n      id\n      displayName\n    }\n  }\n"): (typeof documents)["\n  query stream($id: ID!) {\n    getStreamById(id: $id) {\n      id\n      displayName\n    }\n  }\n"];
/**
 * The gql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function gql(source: "\n  subscription transmissions {\n    allTransmissions {\n      id\n      timestamp\n      downloadUrl\n      transcription {\n        content\n      }\n      chunk {\n        stream {\n          id\n        }\n      }\n    }\n  }\n"): (typeof documents)["\n  subscription transmissions {\n    allTransmissions {\n      id\n      timestamp\n      downloadUrl\n      transcription {\n        content\n      }\n      chunk {\n        stream {\n          id\n        }\n      }\n    }\n  }\n"];
/**
 * The gql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function gql(source: "\n  subscription transcriptions {\n    allTranscriptions {\n      id\n      content\n      transmission {\n        id\n        timestamp\n        downloadUrl\n        chunk {\n          stream {\n            id\n          }\n        }\n      }\n    }\n  }\n"): (typeof documents)["\n  subscription transcriptions {\n    allTranscriptions {\n      id\n      content\n      transmission {\n        id\n        timestamp\n        downloadUrl\n        chunk {\n          stream {\n            id\n          }\n        }\n      }\n    }\n  }\n"];
/**
 * The gql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function gql(source: "\n  query existingTransmissions($stream: ID!, $createdAfter: Time, $after: ID) {\n    getStreamById(id: $stream) {\n      id\n      transmissions(createdAfter: $createdAfter, after: $after) {\n        edges {\n          id\n          timestamp\n          downloadUrl\n          transcription {\n            content\n          }\n        }\n        pageInfo {\n          endCursor\n        }\n      }\n    }\n  }\n"): (typeof documents)["\n  query existingTransmissions($stream: ID!, $createdAfter: Time, $after: ID) {\n    getStreamById(id: $stream) {\n      id\n      transmissions(createdAfter: $createdAfter, after: $after) {\n        edges {\n          id\n          timestamp\n          downloadUrl\n          transcription {\n            content\n          }\n        }\n        pageInfo {\n          endCursor\n        }\n      }\n    }\n  }\n"];
/**
 * The gql function is used to parse GraphQL queries into a document that can be used by GraphQL clients.
 */
export function gql(source: "\n  query GetStreams($after: ID) {\n    getStreams(after: $after) {\n      edges {\n        id\n        createdAt\n        displayName\n        url\n      }\n      pageInfo {\n        endCursor\n      }\n    }\n  }\n"): (typeof documents)["\n  query GetStreams($after: ID) {\n    getStreams(after: $after) {\n      edges {\n        id\n        createdAt\n        displayName\n        url\n      }\n      pageInfo {\n        endCursor\n      }\n    }\n  }\n"];

export function gql(source: string) {
  return (documents as any)[source] ?? {};
}

export type DocumentType<TDocumentNode extends DocumentNode<any, any>> = TDocumentNode extends DocumentNode<  infer TType,  any>  ? TType  : never;