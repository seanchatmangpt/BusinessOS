<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { ArrowLeft, Loader, ExternalLink } from 'lucide-svelte';
  import Button from '$lib/ui/button/Button.svelte';
  import Card from '$lib/ui/card/Card.svelte';
  import ScrollArea from '$lib/ui/scroll-area/ScrollArea.svelte';
  import { getOntologyClass, type OntologyClass } from '$lib/api/ontology';
  import { goto } from '$app/navigation';

  let classDetails: OntologyClass | null = $state(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  const ontologyUri = $derived.by(() => {
    const rawUri = $page.params.ontology;
    return decodeURIComponent(rawUri);
  });

  const className = $derived.by(() => {
    const rawName = $page.params.className;
    return decodeURIComponent(rawName);
  });

  onMount(async () => {
    try {
      classDetails = await getOntologyClass(ontologyUri, className);
      loading = false;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load class details';
      loading = false;
    }
  });

  function getNamespacePrefix(uri: string): string {
    const match = uri.match(/^([^:/]+:)/);
    return match ? match[1] : '';
  }

  function getLocalName(uri: string): string {
    const parts = uri.split(/[/#]/);
    return parts[parts.length - 1];
  }

  function navigateToClass(classUri: string) {
    goto(
      `/ontology/${encodeURIComponent(ontologyUri)}/class/${encodeURIComponent(classUri)}`,
    );
  }
</script>

<div class="w-full min-h-screen bg-gray-50 dark:bg-gray-900">
  <!-- Header with Back Button -->
  <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-6">
    <div class="max-w-5xl mx-auto">
      <div class="flex items-center gap-4 mb-4">
        <Button
          variant="outline"
          size="sm"
          class="gap-2"
          onclick={() => goto(`/ontology/${encodeURIComponent(ontologyUri)}`)}
        >
          <ArrowLeft size={16} />
          Back to Ontology
        </Button>
      </div>
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          {classDetails?.label || classDetails?.name || 'Class'}
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-400 mt-2 font-mono">
          {className}
        </p>
      </div>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center p-12">
      <div class="text-center">
        <Loader class="mx-auto mb-4 animate-spin" size={32} />
        <p class="text-gray-600 dark:text-gray-400">Loading class details...</p>
      </div>
    </div>
  {:else if error}
    <div class="max-w-5xl mx-auto p-6">
      <Card class="p-6 bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800">
        <h3 class="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">Error</h3>
        <p class="text-red-700 dark:text-red-300">{error}</p>
      </Card>
    </div>
  {:else if classDetails}
    <div class="max-w-5xl mx-auto p-6 space-y-6">
      <!-- Class Metadata -->
      <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
        <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Metadata</h2>
        <div class="space-y-4">
          {#if classDetails.comment}
            <div>
              <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
                Definition
              </h3>
              <p class="text-gray-700 dark:text-gray-300">{classDetails.comment}</p>
            </div>
          {/if}

          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <h3 class="font-semibold text-gray-700 dark:text-gray-300 mb-1">Class URI</h3>
              <div class="flex items-center gap-2">
                <code class="text-xs bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded font-mono text-gray-900 dark:text-gray-100 flex-1 truncate">
                  {classDetails.uri}
                </code>
                <button
                  type="button"
                  class="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
                  title="Copy"
                  onclick={() => navigator.clipboard.writeText(classDetails.uri)}
                >
                  <ExternalLink size={14} />
                </button>
              </div>
            </div>

            <div>
              <h3 class="font-semibold text-gray-700 dark:text-gray-300 mb-1">Namespace</h3>
              <code class="text-xs bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded font-mono text-gray-900 dark:text-gray-100">
                {getNamespacePrefix(classDetails.uri)}
              </code>
            </div>
          </div>
        </div>
      </Card>

      <!-- Class Hierarchy -->
      <div class="grid grid-cols-2 gap-6">
        <!-- Parent Classes -->
        {#if classDetails.parentClasses && classDetails.parentClasses.length > 0}
          <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Parent Classes ({classDetails.parentClasses.length})
            </h2>
            <ScrollArea class="max-h-80">
              <div class="space-y-2">
                {#each classDetails.parentClasses as parentUri}
                  <button
                    type="button"
                    class="w-full text-left px-3 py-2 rounded hover:bg-blue-50 dark:hover:bg-blue-900 transition-colors"
                    onclick={() => navigateToClass(parentUri)}
                  >
                    <div class="font-medium text-gray-900 dark:text-white truncate">
                      {getLocalName(parentUri)}
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400 font-mono truncate">
                      {parentUri}
                    </div>
                  </button>
                {/each}
              </div>
            </ScrollArea>
          </Card>
        {/if}

        <!-- Subclasses -->
        {#if classDetails.subClasses && classDetails.subClasses.length > 0}
          <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Subclasses ({classDetails.subClasses.length})
            </h2>
            <ScrollArea class="max-h-80">
              <div class="space-y-2">
                {#each classDetails.subClasses as subUri}
                  <button
                    type="button"
                    class="w-full text-left px-3 py-2 rounded hover:bg-blue-50 dark:hover:bg-blue-900 transition-colors"
                    onclick={() => navigateToClass(subUri)}
                  >
                    <div class="font-medium text-gray-900 dark:text-white truncate">
                      {getLocalName(subUri)}
                    </div>
                    <div class="text-xs text-gray-500 dark:text-gray-400 font-mono truncate">
                      {subUri}
                    </div>
                  </button>
                {/each}
              </div>
            </ScrollArea>
          </Card>
        {/if}
      </div>

      <!-- Datatype Properties -->
      {#if classDetails.dataProperties && classDetails.dataProperties.length > 0}
        <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Datatype Properties ({classDetails.dataProperties.length})
          </h2>
          <div class="space-y-3">
            {#each classDetails.dataProperties as prop}
              <div class="border-l-4 border-blue-500 pl-4 py-2">
                <div class="font-medium text-gray-900 dark:text-white">{prop.label || prop.name}</div>
                {#if prop.comment}
                  <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">{prop.comment}</p>
                {/if}
                <div class="flex flex-wrap gap-2 mt-2">
                  {#if prop.range}
                    <span
                      class="text-xs bg-blue-100 dark:bg-blue-900 text-blue-900 dark:text-blue-100 px-2 py-1 rounded"
                    >
                      Range: {getLocalName(prop.range)}
                    </span>
                  {/if}
                  {#if prop.domain}
                    <span
                      class="text-xs bg-green-100 dark:bg-green-900 text-green-900 dark:text-green-100 px-2 py-1 rounded"
                    >
                      Domain: {getLocalName(prop.domain)}
                    </span>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </Card>
      {/if}

      <!-- Object Properties -->
      {#if classDetails.objectProperties && classDetails.objectProperties.length > 0}
        <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Object Properties ({classDetails.objectProperties.length})
          </h2>
          <div class="space-y-3">
            {#each classDetails.objectProperties as prop}
              <div class="border-l-4 border-purple-500 pl-4 py-2">
                <div class="font-medium text-gray-900 dark:text-white">{prop.label || prop.name}</div>
                {#if prop.comment}
                  <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">{prop.comment}</p>
                {/if}
                <div class="flex flex-wrap gap-2 mt-2">
                  {#if prop.range}
                    <button
                      type="button"
                      class="text-xs bg-purple-100 dark:bg-purple-900 text-purple-900 dark:text-purple-100 px-2 py-1 rounded hover:opacity-80 transition-opacity"
                      onclick={() => navigateToClass(prop.range)}
                    >
                      Range: {getLocalName(prop.range)}
                    </button>
                  {/if}
                  {#if prop.domain}
                    <button
                      type="button"
                      class="text-xs bg-orange-100 dark:bg-orange-900 text-orange-900 dark:text-orange-100 px-2 py-1 rounded hover:opacity-80 transition-opacity"
                      onclick={() => navigateToClass(prop.domain)}
                    >
                      Domain: {getLocalName(prop.domain)}
                    </button>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </Card>
      {/if}
    </div>
  {/if}
</div>
