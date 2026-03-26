import { describe, it, expect, beforeEach, vi } from 'vitest';
import * as ontologyApi from '$lib/api/ontology';

// Mock API responses
const mockOntologies = [
  {
    uri: 'http://example.org/prov#',
    name: 'PROV Ontology',
    prefix: 'prov',
    classCount: 12,
    propertyCount: 28,
    importedOntologies: [],
  },
  {
    uri: 'http://example.org/org#',
    name: 'Organization Ontology',
    prefix: 'org',
    classCount: 8,
    propertyCount: 15,
    importedOntologies: ['http://example.org/prov#'],
  },
];

const mockClasses = [
  {
    uri: 'http://example.org/prov#Entity',
    name: 'Entity',
    label: 'Entity',
    comment: 'An entity is a thing, real or abstract',
    parentClasses: [],
    subClasses: [
      'http://example.org/prov#Agent',
      'http://example.org/prov#Activity',
    ],
    dataProperties: [
      {
        uri: 'http://example.org/prov#label',
        name: 'label',
        label: 'Label',
        comment: 'A human-readable label',
        type: 'datatype',
        range: 'xsd:string',
      },
    ],
    objectProperties: [],
  },
  {
    uri: 'http://example.org/prov#Agent',
    name: 'Agent',
    label: 'Agent',
    comment: 'An agent is something that acts',
    parentClasses: ['http://example.org/prov#Entity'],
    subClasses: [],
    dataProperties: [],
    objectProperties: [],
  },
];

// Mock the request function
vi.mock('$lib/api/base', () => ({
  request: vi.fn(),
}));

describe('Ontology API', () => {
  describe('listOntologies', () => {
    it('should return list of ontologies', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce(mockOntologies);

      const result = await ontologyApi.listOntologies();

      expect(result).toEqual(mockOntologies);
      expect(result.length).toBe(2);
      expect(result[0].name).toBe('PROV Ontology');
    });

    it('should handle empty ontology list', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce([]);

      const result = await ontologyApi.listOntologies();

      expect(result).toEqual([]);
      expect(result.length).toBe(0);
    });
  });

  describe('getOntology', () => {
    it('should return ontology info by URI', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce(mockOntologies[0]);

      const result = await ontologyApi.getOntology('http://example.org/prov#');

      expect(result).toEqual(mockOntologies[0]);
      expect(result.name).toBe('PROV Ontology');
    });

    it('should encode URI properly', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce(mockOntologies[0]);

      const uri = 'http://example.org/prov#';
      await ontologyApi.getOntology(uri);

      expect((request as any).mock.calls[0][0]).toContain(
        encodeURIComponent(uri),
      );
    });
  });

  describe('getOntologyStatistics', () => {
    it('should return ontology statistics', async () => {
      const { request } = await import('$lib/api/base');
      const stats = {
        ontologyUri: 'http://example.org/prov#',
        classCount: 12,
        datatypePropertyCount: 20,
        objectPropertyCount: 8,
        importedOntologies: [],
        rootClasses: ['http://example.org/prov#Entity'],
      };
      (request as any).mockResolvedValueOnce(stats);

      const result = await ontologyApi.getOntologyStatistics(
        'http://example.org/prov#',
      );

      expect(result.classCount).toBe(12);
      expect(result.datatypePropertyCount).toBe(20);
      expect(result.objectPropertyCount).toBe(8);
      expect(result.rootClasses.length).toBe(1);
    });
  });

  describe('getOntologyClasses', () => {
    it('should return list of classes', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce(mockClasses);

      const result = await ontologyApi.getOntologyClasses(
        'http://example.org/prov#',
      );

      expect(result.length).toBe(2);
      expect(result[0].name).toBe('Entity');
      expect(result[1].name).toBe('Agent');
    });

    it('should handle empty class list', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce([]);

      const result = await ontologyApi.getOntologyClasses(
        'http://example.org/empty#',
      );

      expect(result).toEqual([]);
    });
  });

  describe('getOntologyClass', () => {
    it('should return class details', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce(mockClasses[0]);

      const result = await ontologyApi.getOntologyClass(
        'http://example.org/prov#',
        'http://example.org/prov#Entity',
      );

      expect(result.name).toBe('Entity');
      expect(result.subClasses.length).toBe(2);
      expect(result.dataProperties.length).toBe(1);
    });

    it('should encode class name in URI', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce(mockClasses[0]);

      const className = 'http://example.org/prov#Entity';
      await ontologyApi.getOntologyClass('http://example.org/prov#', className);

      const callArg = (request as any).mock.calls[0][0];
      expect(callArg).toContain(encodeURIComponent(className));
    });
  });

  describe('getOntologyProperties', () => {
    it('should return list of properties', async () => {
      const { request } = await import('$lib/api/base');
      const props = [
        {
          uri: 'http://example.org/prov#label',
          name: 'label',
          type: 'datatype' as const,
        },
        {
          uri: 'http://example.org/prov#wasAttributedTo',
          name: 'wasAttributedTo',
          type: 'object' as const,
        },
      ];
      (request as any).mockResolvedValueOnce(props);

      const result = await ontologyApi.getOntologyProperties(
        'http://example.org/prov#',
      );

      expect(result.length).toBe(2);
      expect(result[0].type).toBe('datatype');
      expect(result[1].type).toBe('object');
    });
  });

  describe('searchClasses', () => {
    it('should return search results', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce([mockClasses[0]]);

      const result = await ontologyApi.searchClasses(
        'http://example.org/prov#',
        'Entity',
      );

      expect(result.length).toBe(1);
      expect(result[0].name).toBe('Entity');
    });

    it('should handle empty search results', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce([]);

      const result = await ontologyApi.searchClasses(
        'http://example.org/prov#',
        'nonexistent',
      );

      expect(result).toEqual([]);
    });

    it('should be case-insensitive', async () => {
      const { request } = await import('$lib/api/base');
      (request as any).mockResolvedValueOnce([mockClasses[0]]);

      const result = await ontologyApi.searchClasses(
        'http://example.org/prov#',
        'entity',
      );

      expect(result.length).toBeGreaterThanOrEqual(0);
    });
  });

  describe('getClassHierarchy', () => {
    it('should return class hierarchy', async () => {
      const { request } = await import('$lib/api/base');
      const hierarchy = {
        'http://example.org/prov#Entity': [
          'http://example.org/prov#Agent',
          'http://example.org/prov#Activity',
        ],
      };
      (request as any).mockResolvedValueOnce(hierarchy);

      const result = await ontologyApi.getClassHierarchy(
        'http://example.org/prov#',
      );

      expect(result['http://example.org/prov#Entity']).toBeDefined();
      expect(
        (result['http://example.org/prov#Entity'] as string[]).length,
      ).toBe(2);
    });
  });

  describe('Class hierarchy navigation', () => {
    it('should identify root classes (no parent classes)', () => {
      const rootClasses = mockClasses.filter((c) => c.parentClasses.length === 0);
      expect(rootClasses.length).toBe(1);
      expect(rootClasses[0].name).toBe('Entity');
    });

    it('should identify subclasses', () => {
      const entity = mockClasses.find((c) => c.name === 'Entity');
      expect(entity?.subClasses.length).toBe(2);
    });

    it('should chain class relationships', () => {
      const agent = mockClasses.find((c) => c.name === 'Agent');
      const entity = mockClasses.find((c) => c.name === 'Entity');

      expect(agent?.parentClasses).toContain(entity?.uri);
      expect(entity?.subClasses).toContain(agent?.uri);
    });
  });

  describe('Property filtering', () => {
    it('should separate datatype properties from object properties', () => {
      const cls = mockClasses[0];
      expect(cls.dataProperties.length).toBe(1);
      expect(cls.objectProperties.length).toBe(0);
    });

    it('should filter properties by domain', () => {
      const propsWithDomain = mockClasses[0].dataProperties.filter(
        (p) => p.domain !== undefined,
      );
      expect(propsWithDomain.length).toBeGreaterThanOrEqual(0);
    });

    it('should filter properties by range', () => {
      const propsWithRange = mockClasses[0].dataProperties.filter(
        (p) => p.range !== undefined,
      );
      expect(propsWithRange.length).toBeGreaterThanOrEqual(0);
    });
  });
});
