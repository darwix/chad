<script>
	import { onMount } from 'svelte';
	import Chat from './lib/Chat.svelte';
	import Login from './lib/Login.svelte';
	import { supabase } from './lib/supabase';

	let session = $state(null);
	let loading = $state(true);
	let darkMode = $state(false);

	onMount(() => {
		const savedTheme = localStorage.getItem('theme');
		if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
			darkMode = true;
		}

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

	$effect(() => {
		if (darkMode) {
			document.documentElement.classList.add('dark');
			localStorage.setItem('theme', 'dark');
		} else {
			document.documentElement.classList.remove('dark');
			localStorage.setItem('theme', 'light');
		}
	});

	async function handleLogout() {
		await supabase.auth.signOut();
	}

	function toggleDarkMode() {
		darkMode = !darkMode;
	}

	const username = $derived(session?.user?.email?.split('@')[0] || 'Unknown');
</script>

<main class="h-screen w-full bg-gray-100 dark:bg-gray-900 dark:text-gray-100 transition-colors duration-200">
	<div class="fixed top-2 right-2 md:top-4 md:right-4 z-40 flex items-center gap-2">
		<button
			onclick={toggleDarkMode}
			class="p-2 rounded-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 shadow-sm hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
			aria-label="Toggle dark mode"
		>
			{#if darkMode}
				<svg class="w-4 h-4 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
					<path d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" />
				</svg>
			{:else}
				<svg class="w-4 h-4 text-gray-600" fill="currentColor" viewBox="0 0 20 20">
					<path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z" />
				</svg>
			{/if}
		</button>
		{#if session}
			<button
				onclick={handleLogout}
				class="text-xs font-medium text-gray-600 dark:text-gray-300 hover:text-red-600 dark:hover:text-red-400 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 px-3 py-1.5 rounded-full shadow-sm"
			>
				Logout
			</button>
		{/if}
	</div>

	{#if loading}
		<div class="flex items-center justify-center h-full">
			<p>Loading...</p>
		</div>
	{:else if !session}
		<Login />
	{:else}
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
