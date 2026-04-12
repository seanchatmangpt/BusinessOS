<script lang="ts">
  import ClassTreeNode from './ClassTreeNode.svelte';
  import type { OntologyClass } from '$lib/api/ontology';

  interface Props {
    classes: OntologyClass[];
    selectedClass?: OntologyClass;
    onSelect: (cls: OntologyClass) => void;
    rootClasses?: string[];
  }

  let { classes, selectedClass, onSelect, rootClasses = [] }: Props = $props();

  let expandedNodes = $state<Set<string>>(new Set());

  function toggleNode(classUri: string) {
    if (expandedNodes.has(classUri)) {
      expandedNodes.delete(classUri);
    } else {
      expandedNodes.add(classUri);
    }
    expandedNodes = new Set(expandedNodes);
  }

  function isExpanded(classUri: string): boolean {
    return expandedNodes.has(classUri);
  }

  function isRootClass(cls: OntologyClass): boolean {
    if (rootClasses.length === 0) return true;
    return rootClasses.includes(cls.uri);
  }

  function getSubClasses(parentUri: string): OntologyClass[] {
    return classes.filter((c) => c.parentClasses?.includes(parentUri));
  }
</script>

<div class="class-tree space-y-1 text-sm font-mono">
  {#each classes.filter((c) => isRootClass(c)) as cls (cls.uri)}
    <ClassTreeNode
      {cls}
      {selectedClass}
      {onSelect}
      onToggle={toggleNode}
      {isExpanded}
      {getSubClasses}
    />
  {/each}
</div>

<style>
  :global(.class-tree) {
    user-select: none;
  }
</style>
