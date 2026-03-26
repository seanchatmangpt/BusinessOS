import { request } from './base';

export interface OntologyInfo {
  uri: string;
  name: string;
  prefix: string;
  classCount: number;
  propertyCount: number;
  importedOntologies: string[];
}

export interface OntologyClass {
  uri: string;
  name: string;
  label?: string;
  comment?: string;
  parentClasses: string[];
  subClasses: string[];
  dataProperties: OntologyProperty[];
  objectProperties: OntologyProperty[];
}

export interface OntologyProperty {
  uri: string;
  name: string;
  label?: string;
  comment?: string;
  domain?: string;
  range?: string;
  type: 'datatype' | 'object';
}

export interface OntologyStatistics {
  ontologyUri: string;
  classCount: number;
  datatypePropertyCount: number;
  objectPropertyCount: number;
  importedOntologies: string[];
  rootClasses: string[];
}

export async function listOntologies(): Promise<OntologyInfo[]> {
  return request<OntologyInfo[]>('/ontology/list');
}

export async function getOntology(ontologyUri: string): Promise<OntologyInfo> {
  const encoded = encodeURIComponent(ontologyUri);
  return request<OntologyInfo>(`/ontology/${encoded}`);
}

export async function getOntologyStatistics(ontologyUri: string): Promise<OntologyStatistics> {
  const encoded = encodeURIComponent(ontologyUri);
  return request<OntologyStatistics>(`/ontology/${encoded}/statistics`);
}

export async function getOntologyClasses(ontologyUri: string): Promise<OntologyClass[]> {
  const encoded = encodeURIComponent(ontologyUri);
  return request<OntologyClass[]>(`/ontology/${encoded}/classes`);
}

export async function getOntologyClass(
  ontologyUri: string,
  className: string,
): Promise<OntologyClass> {
  const ontologyEncoded = encodeURIComponent(ontologyUri);
  const classEncoded = encodeURIComponent(className);
  return request<OntologyClass>(`/ontology/${ontologyEncoded}/class/${classEncoded}`);
}

export async function getOntologyProperties(ontologyUri: string): Promise<OntologyProperty[]> {
  const encoded = encodeURIComponent(ontologyUri);
  return request<OntologyProperty[]>(`/ontology/${encoded}/properties`);
}

export async function searchClasses(
  ontologyUri: string,
  query: string,
): Promise<OntologyClass[]> {
  const encoded = encodeURIComponent(ontologyUri);
  const params = new URLSearchParams({ q: query });
  return request<OntologyClass[]>(`/ontology/${encoded}/search?${params}`);
}

export async function getClassHierarchy(ontologyUri: string): Promise<Record<string, unknown>> {
  const encoded = encodeURIComponent(ontologyUri);
  return request<Record<string, unknown>>(`/ontology/${encoded}/hierarchy`);
}
