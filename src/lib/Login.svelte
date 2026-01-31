<script>
	import { supabase } from './supabase';

	let email = $state('');
	let password = $state('');
	let isRegistering = $state(false);
	let loading = $state(false);
	let message = $state('');
	let messageType = $state('info'); // 'info' or 'error'

	async function handleAuth(e) {
		e.preventDefault();
		loading = true;
		message = '';

		if (isRegistering) {
			const { error } = await supabase.auth.signUp({
				email,
				password
			});
			if (error) {
				message = error.message;
				messageType = 'error';
			} else {
				message = 'Registration successful! You can now log in.';
				messageType = 'info';
				isRegistering = false;
			}
		} else {
			const { error } = await supabase.auth.signInWithPassword({
				email,
				password
			});
			if (error) {
				message = error.message;
				messageType = 'error';
			}
		}
		loading = false;
	}
</script>

<div class="flex flex-col items-center justify-center h-full space-y-4">
	<div class="bg-white p-8 rounded-lg shadow-md w-96">
		<h1 class="text-2xl font-bold mb-4 text-center">
			{isRegistering ? 'Create Account' : 'Login'}
		</h1>

		{#if message}
			<div
				class="mb-4 p-3 rounded text-sm {messageType === 'error'
					? 'bg-red-50 text-red-700'
					: 'bg-blue-50 text-blue-700'}"
			>
				{message}
			</div>
		{/if}

		<form onsubmit={handleAuth} class="space-y-4">
			<div>
				<label for="email" class="block text-sm font-medium text-gray-700">Email Address</label>
				<input
					type="email"
					id="email"
					bind:value={email}
					class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-sm"
					placeholder="your@email.com"
					required
					disabled={loading}
				/>
			</div>
			<div>
				<label for="password" class="block text-sm font-medium text-gray-700">Password</label>
				<input
					type="password"
					id="password"
					bind:value={password}
					class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-sm"
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
				{loading ? 'Processing...' : isRegistering ? 'Sign Up' : 'Login'}
			</button>
		</form>

		<div class="mt-6 text-center">
			<button
				onclick={() => {
					isRegistering = !isRegistering;
					message = '';
				}}
				class="text-sm text-blue-600 hover:underline focus:outline-none"
			>
				{isRegistering ? 'Already have an account? Login' : "Don't have an account? Sign Up"}
			</button>
		</div>
	</div>
</div>
