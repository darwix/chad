<script>
	import { onMount } from 'svelte';
	import { getLinkPreview } from 'link-preview-js';

	let { url } = $props();

	let metadata = $state(null);
	let loading = $state(true);
	let error = $state(false);

	onMount(async () => {
		try {
			// Using a CORS proxy to allow the library to fetch metadata in a frontend-only environment
			const proxyUrl = `https://corsproxy.io/?${encodeURIComponent(url)}`;
			const data = await getLinkPreview(proxyUrl);
			metadata = {
				title: data.title,
				description: data.description,
				image: data.images && data.images.length > 0 ? { url: data.images[0] } : null
			};
		} catch (e) {
			console.error('Error fetching link preview:', e);
			error = true;
		} finally {
			loading = false;
		}
	});
</script>

{#if !error}
	<div class="mt-2 bg-white dark:bg-gray-700 rounded-lg border dark:border-gray-600 overflow-hidden shadow-sm max-w-sm">
		{#if loading}
			<div class="p-4 flex items-center space-x-2 animate-pulse">
				<div class="w-12 h-12 bg-gray-200 dark:bg-gray-600 rounded"></div>
				<div class="flex-1 space-y-2">
					<div class="h-3 bg-gray-200 dark:bg-gray-600 rounded w-3/4"></div>
					<div class="h-2 bg-gray-200 dark:bg-gray-600 rounded w-1/2"></div>
				</div>
			</div>
		{:else if metadata}
			<a href={url} target="_blank" rel="noopener noreferrer" class="block hover:bg-gray-50 dark:hover:bg-gray-600/50 transition-colors">
				{#if metadata.image?.url}
					<img src={metadata.image.url} alt={metadata.title} class="w-full h-40 object-cover" />
				{/if}
				<div class="p-3">
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white truncate">{metadata.title || url}</h3>
					{#if metadata.description}
						<p class="mt-1 text-xs text-gray-500 dark:text-gray-400 line-clamp-2">{metadata.description}</p>
					{/if}
					<p class="mt-1 text-[10px] text-gray-400 dark:text-gray-500 truncate">
						{(() => {
							try {
								return new URL(url).hostname;
							} catch {
								return '';
							}
						})()}
					</p>
				</div>
			</a>
		{/if}
	</div>
{/if}
