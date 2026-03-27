<script lang="ts">
  import { ChevronRight, ChevronDown } from 'lucide-svelte';
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

  function getClassByUri(uri: string): OntologyClass | undefined {
    return classes.find((c) => c.uri === uri);
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
      {onToggle: toggleNode}
      {isExpanded}
      {getSubClasses}
      {getClassByUri}
    />
  {/each}
</div>

<style>
  :global(.class-tree) {
    user-select: none;
  }
</style>

<script lang="ts">
  interface ClassTreeNodeProps {
    cls: OntologyClass;
    selectedClass?: OntologyClass;
    onSelect: (cls: OntologyClass) => void;
    onToggle: (uri: string) => void;
    isExpanded: (uri: string) => boolean;
    getSubClasses: (uri: string) => OntologyClass[];
    getClassByUri: (uri: string) => OntologyClass | undefined;
    level?: number;
  }

  function ClassTreeNode({
    cls,
    selectedClass,
    onSelect,
    onToggle,
    isExpanded,
    getSubClasses,
    level = 0,
  }: ClassTreeNodeProps) {
    const subClasses = getSubClasses(cls.uri);
    const hasChildren = subClasses.length > 0;
    const expanded = isExpanded(cls.uri);
    const isSelected = selectedClass?.uri === cls.uri;

    return (
      <div key={cls.uri} class="tree-node-container">
        <div
          class={`tree-node flex items-center gap-1 px-2 py-1 rounded cursor-pointer transition-colors ${
            isSelected
              ? 'bg-blue-100 text-blue-900 dark:bg-blue-900 dark:text-blue-100'
              : 'hover:bg-gray-100 dark:hover:bg-gray-800'
          }`}
          style={`margin-left: ${level * 1.5}rem`}
        >
          {#if hasChildren}
            <button
              type="button"
              class="toggle-btn p-0 w-5 h-5 flex items-center justify-center hover:bg-gray-300 dark:hover:bg-gray-700 rounded"
              onclick={(e) => {
                e.stopPropagation();
                onToggle(cls.uri);
              }}
            >
              {#if expanded}
                <ChevronDown size={16} />
              {:else}
                <ChevronRight size={16} />
              {/if}
            </button>
          {:else}
            <div class="w-5"></div>
          {/if}
          <button
            type="button"
            class="text-left flex-1 truncate hover:underline"
            onclick={() => onSelect(cls)}
          >
            {cls.label || cls.name}
          </button>
        </div>

        {#if expanded && hasChildren}
          <div class="tree-children">
            {#each subClasses as subCls (subCls.uri)}
              <ClassTreeNode
                cls={subCls}
                {selectedClass}
                {onSelect}
                {onToggle}
                {isExpanded}
                {getSubClasses}
                getClassByUri={() => subCls}
                level={level + 1}
              />
            {/each}
          </div>
        {/if}
      </div>
    );
  }
</script>
