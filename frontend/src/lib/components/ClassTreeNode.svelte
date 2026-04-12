<script lang="ts">
  import { ChevronRight, ChevronDown } from 'lucide-svelte';
  import type { OntologyClass } from '$lib/api/ontology';

  interface Props {
    cls: OntologyClass;
    selectedClass?: OntologyClass;
    onSelect: (cls: OntologyClass) => void;
    onToggle: (uri: string) => void;
    isExpanded: (uri: string) => boolean;
    getSubClasses: (uri: string) => OntologyClass[];
    level?: number;
  }

  let { cls, selectedClass, onSelect, onToggle, isExpanded, getSubClasses, level = 0 }: Props = $props();

  const subClasses = $derived(getSubClasses(cls.uri));
  const hasChildren = $derived(subClasses.length > 0);
  const expanded = $derived(isExpanded(cls.uri));
  const isSelected = $derived(selectedClass?.uri === cls.uri);
</script>

<div class="tree-node-container">
  <div
    class="tree-node flex items-center gap-1 px-2 py-1 rounded cursor-pointer transition-colors {isSelected
      ? 'bg-blue-100 text-blue-900 dark:bg-blue-900 dark:text-blue-100'
      : 'hover:bg-gray-100 dark:hover:bg-gray-800'}"
    style="margin-left: {level * 1.5}rem"
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
        <svelte:self
          cls={subCls}
          {selectedClass}
          {onSelect}
          {onToggle}
          {isExpanded}
          {getSubClasses}
          level={level + 1}
        />
      {/each}
    </div>
  {/if}
</div>
