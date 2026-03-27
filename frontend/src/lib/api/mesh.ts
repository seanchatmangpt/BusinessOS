import { getApiBaseUrl } from './base';

export interface Domain {
  id: string;
  name: string;
  owner: string;
  governance_model: string;
  sla: string;
  created_at: string;
  updated_at: string;
}

export interface Dataset {
  id: string;
  domain_id: string;
  name: string;
  owner: string;
  quality_score: number;
  last_modified: string;
  created_at: string;
  updated_at: string;
}

export interface QualityMetrics {
  dataset_id: string;
  completeness: number; // 0-100
  accuracy: number; // 0-100
  consistency: number; // 0-100
  timeliness: number; // 0-100
  overall: number; // 0-100
  last_calculated: string;
}

export interface LineageNode {
  id: string;
  dataset_id: string;
  dataset_name: string;
  quality_score: number;
  level: number; // depth from source
}

export interface LineageEdge {
  source_id: string;
  target_id: string;
  relationship: string;
}

export interface Lineage {
  nodes: LineageNode[];
  edges: LineageEdge[];
  max_depth: number;
}

export interface DataContract {
  id: string;
  dataset_id: string;
  name: string;
  constraints: Array<{
    field: string;
    rule: string;
    severity: 'warn' | 'error';
  }>;
  created_at: string;
  updated_at: string;
}

export class MeshApiClient {
  private baseUrl: string;

  constructor() {
    this.baseUrl = getApiBaseUrl();
  }

  async listDomains(): Promise<Domain[]> {
    const response = await fetch(`${this.baseUrl}/mesh/domains`);
    if (!response.ok) throw new Error('Failed to list domains');
    const data = await response.json();
    return data.domains || [];
  }

  async getDatasets(domainId: string): Promise<Dataset[]> {
    const response = await fetch(`${this.baseUrl}/mesh/domains/${domainId}/datasets`);
    if (!response.ok) throw new Error('Failed to get datasets');
    const data = await response.json();
    return data.datasets || [];
  }

  async getQuality(datasetId: string): Promise<QualityMetrics> {
    const response = await fetch(`${this.baseUrl}/mesh/datasets/${datasetId}/quality`);
    if (!response.ok) throw new Error('Failed to get quality metrics');
    return await response.json();
  }

  async getLineage(datasetId: string, maxDepth: number = 5): Promise<Lineage> {
    const params = new URLSearchParams({ max_depth: maxDepth.toString() });
    const response = await fetch(`${this.baseUrl}/mesh/datasets/${datasetId}/lineage?${params}`);
    if (!response.ok) throw new Error('Failed to get lineage');
    return await response.json();
  }

  async getContracts(datasetId: string): Promise<DataContract[]> {
    const response = await fetch(`${this.baseUrl}/mesh/datasets/${datasetId}/contracts`);
    if (!response.ok) throw new Error('Failed to get contracts');
    const data = await response.json();
    return data.contracts || [];
  }

  async getDomain(domainId: string): Promise<Domain> {
    const response = await fetch(`${this.baseUrl}/mesh/domains/${domainId}`);
    if (!response.ok) throw new Error('Failed to get domain');
    return await response.json();
  }
}

export const meshClient = new MeshApiClient();
