<script>
	import { onMount } from 'svelte';
	import Chat from './lib/Chat.svelte';
	import Login from './lib/Login.svelte';
	import { supabase } from './lib/supabase';

	let session = $state(null);
	let loading = $state(true);

	onMount(() => {
		supabase.auth.getSession().then(({ data: { session: currentSession } }) => {
			session = currentSession;
			loading = false;
		});

		const {
			data: { subscription }
		} = supabase.auth.onAuthStateChange((_event, currentSession) => {
			session = currentSession;
			loading = false;
		});

		return () => subscription.unsubscribe();
	});

	async function handleLogout() {
		await supabase.auth.signOut();
	}

	const username = $derived(session?.user?.email?.split('@')[0] || 'Unknown');
</script>

<main class="h-screen w-full bg-gray-100">
	{#if loading}
		<div class="flex items-center justify-center h-full">
			<p>Loading...</p>
		</div>
	{:else if !session}
		<Login />
	{:else}
		<div class="fixed top-2 right-2 md:top-4 md:right-4 z-40">
			<button
				onclick={handleLogout}
				class="text-xs font-medium text-gray-600 hover:text-red-600 bg-white border border-gray-200 px-3 py-1.5 rounded-full shadow-sm"
			>
				Logout
			</button>
		</div>
		<Chat {username} />
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		padding: 0;
		font-family:
			-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
	}
</style>
