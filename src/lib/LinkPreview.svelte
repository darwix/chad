<script>
	import { onMount } from 'svelte';
	import { getLinkPreview } from 'link-preview-js';

	let { url } = $props();

	let metadata = $state(null);
	let loading = $state(true);
	let error = $state(false);

	const difficultDomains = ['archiveofourown.org'];

	async function fetchWithMicrolink() {
		const response = await fetch(`https://api.microlink.io/?url=${encodeURIComponent(url)}`);
		const result = await response.json();

		if (result.status === 'success' && result.data.title) {
			metadata = {
				title: result.data.title,
				description: result.data.description,
				image: result.data.image ? { url: result.data.image.url } : null
			};
			return true;
		}
		return false;
	}

	onMount(async () => {
		try {
			const hostname = new URL(url).hostname;
			const isDifficult = difficultDomains.some((d) => hostname.includes(d));

			if (isDifficult) {
				const success = await fetchWithMicrolink();
				if (success) return;
			}

			// Try primary method: link-preview-js with corsproxy.io
			const proxyUrl = `https://corsproxy.io/?${encodeURIComponent(url)}`;
			const data = await getLinkPreview(proxyUrl);

			// Check if the content is a bot detection page (e.g. Cloudflare)
			const isBotBlock =
				(data.title &&
					(data.title.includes('Cloudflare') ||
						data.title.includes('Attention Required') ||
						data.title.includes('Just a moment'))) ||
				(data.description && data.description.includes('Cloudflare'));

			if (isBotBlock) {
				throw new Error('Bot block detected');
			}

			metadata = {
				title: data.title,
				description: data.description,
				image: data.images && data.images.length > 0 ? { url: data.images[0] } : null
			};
		} catch (e) {
			console.log('Primary preview method failed or skipped, trying workaround...', e.message);
			try {
				if (!metadata) {
					const success = await fetchWithMicrolink();
					if (!success) error = true;
				}
			} catch (fallbackError) {
				console.error('Workaround failed too:', fallbackError);
				error = true;
			}
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
