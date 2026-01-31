<script>
	let { onlogin } = $props();
	let name = $state('');

	function handleSubmit(e) {
		e.preventDefault();
		if (name.trim()) {
			localStorage.setItem('chat_username', name);
			onlogin(name);
		}
	}
</script>

<div class="flex flex-col items-center justify-center h-full space-y-4">
	<div class="bg-white p-8 rounded-lg shadow-md w-96">
		<h1 class="text-2xl font-bold mb-4 text-center">Join Chat</h1>
		<form onsubmit={handleSubmit} class="space-y-4">
			<div>
				<label for="name" class="block text-sm font-medium text-gray-700">Enter your name</label>
				<input
					type="text"
					id="name"
					bind:value={name}
					class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
					placeholder="Your name..."
					required
				/>
			</div>
			<button
				type="submit"
				class="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
			>
				Join
			</button>
		</form>
	</div>
</div>
