<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { ArrowLeft, Loader } from 'lucide-svelte';
  import Button from '$lib/ui/button/Button.svelte';
  import Card from '$lib/ui/card/Card.svelte';
  import ScrollArea from '$lib/ui/scroll-area/ScrollArea.svelte';
  import ClassTree from '$lib/components/ClassTree.svelte';
  import {
    getOntology,
    getOntologyStatistics,
    getOntologyClasses,
    type OntologyInfo,
    type OntologyStatistics,
    type OntologyClass,
  } from '$lib/api/ontology';
  import { goto } from '$app/navigation';

  let ontologyInfo: OntologyInfo | null = $state(null);
  let statistics: OntologyStatistics | null = $state(null);
  let classes: OntologyClass[] = $state([]);
  let selectedClass: OntologyClass | null = $state(null);
  let loading = $state(true);
  let error = $state<string | null>(null);

  const ontologyUri = $derived.by(() => {
    const rawUri = $page.params.ontology;
    return decodeURIComponent(rawUri);
  });

  onMount(async () => {
    try {
      const [onto, stats, classList] = await Promise.all([
        getOntology(ontologyUri),
        getOntologyStatistics(ontologyUri),
        getOntologyClasses(ontologyUri),
      ]);

      ontologyInfo = onto;
      statistics = stats;
      classes = classList;
      loading = false;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load ontology';
      loading = false;
    }
  });

  function handleSelectClass(cls: OntologyClass) {
    selectedClass = cls;
    goto(`/ontology/${encodeURIComponent(ontologyUri)}/class/${encodeURIComponent(cls.uri)}`);
  }
</script>

<div class="w-full h-screen flex flex-col bg-gray-50 dark:bg-gray-900">
  <!-- Header with Back Button -->
  <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-6">
    <div class="max-w-7xl mx-auto flex items-center gap-4">
      <Button
        variant="outline"
        size="sm"
        class="gap-2"
        onclick={() => goto('/ontology')}
      >
        <ArrowLeft size={16} />
        Back
      </Button>
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
          {ontologyInfo?.name || 'Ontology'}
        </h1>
        <p class="text-sm text-gray-600 dark:text-gray-400 mt-1 font-mono">
          {ontologyUri}
        </p>
      </div>
    </div>
  </div>

  {#if loading}
    <div class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <Loader class="mx-auto mb-4 animate-spin" size={32} />
        <p class="text-gray-600 dark:text-gray-400">Loading ontology details...</p>
      </div>
    </div>
  {:else if error}
    <div class="flex-1 flex items-center justify-center p-6">
      <Card class="max-w-md p-6 bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800">
        <h3 class="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">Error</h3>
        <p class="text-red-700 dark:text-red-300">{error}</p>
      </Card>
    </div>
  {:else}
    <div class="flex-1 flex overflow-hidden">
      <!-- Sidebar: Class Tree -->
      <div class="w-80 border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 flex flex-col">
        <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            Classes
          </h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            {classes.length} total
          </p>
        </div>

        <ScrollArea class="flex-1">
          <div class="p-4">
            <ClassTree
              {classes}
              selectedClass={selectedClass}
              onSelect={handleSelectClass}
              rootClasses={statistics?.rootClasses || []}
            />
          </div>
        </ScrollArea>
      </div>

      <!-- Main Content: Statistics and Details -->
      <div class="flex-1 flex flex-col overflow-hidden p-6">
        <div class="space-y-6 overflow-y-auto">
          <!-- Statistics Cards -->
          {#if statistics}
            <div class="grid grid-cols-3 gap-4">
              <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
                <div class="text-3xl font-bold text-blue-600 dark:text-blue-400">
                  {statistics.classCount}
                </div>
                <p class="text-sm text-gray-600 dark:text-gray-400 mt-2">Classes</p>
              </Card>

              <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
                <div class="text-3xl font-bold text-green-600 dark:text-green-400">
                  {statistics.datatypePropertyCount + statistics.objectPropertyCount}
                </div>
                <p class="text-sm text-gray-600 dark:text-gray-400 mt-2">Properties</p>
              </Card>

              <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
                <div class="text-3xl font-bold text-purple-600 dark:text-purple-400">
                  {statistics.importedOntologies.length}
                </div>
                <p class="text-sm text-gray-600 dark:text-gray-400 mt-2">Imports</p>
              </Card>
            </div>
          {/if}

          <!-- Property Breakdown -->
          {#if statistics}
            <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                Property Types
              </h3>
              <div class="space-y-3">
                <div class="flex justify-between">
                  <span class="text-gray-700 dark:text-gray-300">Datatype Properties</span>
                  <span class="font-medium text-gray-900 dark:text-white">
                    {statistics.datatypePropertyCount}
                  </span>
                </div>
                <div class="flex justify-between">
                  <span class="text-gray-700 dark:text-gray-300">Object Properties</span>
                  <span class="font-medium text-gray-900 dark:text-white">
                    {statistics.objectPropertyCount}
                  </span>
                </div>
              </div>
            </Card>
          {/if}

          <!-- Imported Ontologies -->
          {#if statistics && statistics.importedOntologies.length > 0}
            <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                Imported Ontologies
              </h3>
              <div class="space-y-2">
                {#each statistics.importedOntologies as imported}
                  <div
                    class="flex items-center gap-2 px-3 py-2 bg-gray-100 dark:bg-gray-700 rounded text-sm text-gray-800 dark:text-gray-200"
                  >
                    <span class="truncate font-mono" title={imported}>
                      {imported}
                    </span>
                  </div>
                {/each}
              </div>
            </Card>
          {/if}

          <!-- Root Classes -->
          {#if statistics && statistics.rootClasses.length > 0}
            <Card class="p-6 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                Root Classes
              </h3>
              <div class="space-y-1">
                {#each statistics.rootClasses as rootUri}
                  {@const rootCls = classes.find((c) => c.uri === rootUri)}
                  {#if rootCls}
                    <button
                      type="button"
                      class="w-full text-left px-3 py-2 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 transition-colors"
                      onclick={() => handleSelectClass(rootCls)}
                    >
                      <div class="font-medium truncate">{rootCls.label || rootCls.name}</div>
                      <div class="text-xs text-gray-500 dark:text-gray-400 font-mono truncate">
                        {rootUri}
                      </div>
                    </button>
                  {/if}
                {/each}
              </div>
            </Card>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>
