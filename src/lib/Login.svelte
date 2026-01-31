<script>
	import { supabase } from './supabase';

	let email = $state('');
	let password = $state('');
	let loading = $state(false);
	let message = $state('');
	let messageType = $state('info'); // 'info' or 'error'

	async function handleAuth(e) {
		e.preventDefault();
		loading = true;
		message = '';

		const { error } = await supabase.auth.signInWithPassword({
			email,
			password
		});
		
		if (error) {
			message = error.message;
			messageType = 'error';
		}
		
		loading = false;
	}
</script>

<div class="flex flex-col items-center justify-center min-h-screen p-4 bg-gray-100 dark:bg-gray-900 transition-colors duration-200">
	<div class="bg-white dark:bg-gray-800 p-6 md:p-8 rounded-lg shadow-md w-full max-w-sm">
		<h1 class="text-2xl font-bold mb-6 text-center text-gray-800 dark:text-white">Login</h1>

		{#if message}
			<div
				class="mb-4 p-3 rounded text-sm {messageType === 'error'
					? 'bg-red-50 text-red-700 dark:bg-red-900/30 dark:text-red-400'
					: 'bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'}"
			>
				{message}
			</div>
		{/if}

		<form onsubmit={handleAuth} class="space-y-4">
			<div>
				<label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Email Address</label>
				<input
					type="email"
					id="email"
					bind:value={email}
					class="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-sm"
					placeholder="your@email.com"
					required
					disabled={loading}
				/>
			</div>
			<div>
				<label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Password</label>
				<input
					type="password"
					id="password"
					bind:value={password}
					class="mt-1 block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-sm"
					placeholder="••••••••"
					required
					disabled={loading}
				/>
			</div>
			<button
				type="submit"
				class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
				disabled={loading}
			>
				{loading ? 'Logging in...' : 'Login'}
			</button>
		</form>
	</div>
</div>
