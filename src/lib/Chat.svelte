<script>
	import { onMount } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { supabase } from './supabase';

	let { username } = $props();

	let messages = $state([]);
	let newMessage = $state('');
	let onlineUsers = $state([]);
	let typingUsers = new SvelteSet();
	let typingTimeout;
	let chatContainer;
	let channel;
	let isSidebarOpen = $state(false);

	const emojis = ['😀', '😂', '😍', '👍', '🔥', '🎉', '❤️', '🤔', '🙌', '✨'];

	async function fetchHistory() {
		const { data, error } = await supabase
			.from('messages')
			.select('*')
			.order('created_at', { ascending: false })
			.limit(50);

		if (error) {
			console.error('Error fetching history:', error);
		} else {
			messages = data.reverse();
		}
	}

	function setupSupabase() {
		channel = supabase.channel('global-chat', {
			config: {
				presence: {
					key: username
				}
			}
		});

		// Listen for new messages via Postgres Changes
		channel
			.on(
				'postgres_changes',
				{ event: 'INSERT', schema: 'public', table: 'messages' },
				(payload) => {
					messages = [...messages, payload.new];
				}
			)
			// Listen for typing indicators via Broadcast
			.on('broadcast', { event: 'typing' }, (payload) => {
				if (payload.payload.sender !== username) {
					if (payload.payload.is_typing) {
						typingUsers.add(payload.payload.sender);
					} else {
						typingUsers.delete(payload.payload.sender);
					}
				}
			})
			// Track Presence (Online Users)
			.on('presence', { event: 'sync' }, () => {
				const state = channel.presenceState();
				onlineUsers = Object.keys(state);
			})
			.subscribe(async (status) => {
				if (status === 'SUBSCRIBED') {
					await channel.track({ online_at: new Date().toISOString() });
				}
			});
	}

	onMount(() => {
		fetchHistory();
		setupSupabase();
		return () => {
			if (channel) supabase.removeChannel(channel);
		};
	});

	$effect(() => {
		// Auto-scroll when messages change
		messages;
		if (chatContainer) {
			chatContainer.scrollTop = chatContainer.scrollHeight;
		}
	});

	async function sendMessage() {
		if (newMessage.trim()) {
			const msg = {
				sender: username,
				content: newMessage
			};

			const { error } = await supabase.from('messages').insert([msg]);
			if (error) {
				console.error('Error sending message:', error);
			} else {
				newMessage = '';
				sendTyping(false);
			}
		}
	}

	function handleKeydown(e) {
		if (e.key === 'Enter') {
			sendMessage();
		} else {
			sendTyping(true);
			clearTimeout(typingTimeout);
			typingTimeout = setTimeout(() => sendTyping(false), 2000);
		}
	}

	function sendTyping(isTyping) {
		if (channel) {
			channel.send({
				type: 'broadcast',
				event: 'typing',
				payload: { sender: username, is_typing: isTyping }
			});
		}
	}

	function addEmoji(emoji) {
		newMessage += emoji;
	}

	function formatTime(ts) {
		return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}

	function toggleSidebar() {
		isSidebarOpen = !isSidebarOpen;
	}
</script>

<div class="flex h-screen bg-gray-100 overflow-hidden relative">
	<!-- Sidebar Overlay -->
	{#if isSidebarOpen}
		<button
			class="fixed inset-0 bg-black/50 z-20 md:hidden"
			onclick={toggleSidebar}
			aria-label="Close sidebar"
		></button>
	{/if}

	<!-- Sidebar -->
	<aside
		class="absolute md:relative w-64 h-full bg-white border-r flex flex-col z-30 transition-transform duration-300 transform {isSidebarOpen
			? 'translate-x-0'
			: '-translate-x-full md:translate-x-0'}"
	>
		<div class="p-4 border-b font-bold text-lg flex justify-between items-center">
			<span>Online Users ({onlineUsers.length})</span>
			<button class="md:hidden text-gray-500" onclick={toggleSidebar}>
				✕
			</button>
		</div>
		<div class="flex-1 overflow-y-auto p-4 space-y-2">
			{#each onlineUsers as user (user)}
				<div class="flex items-center space-x-2">
					<div class="w-2 h-2 bg-green-500 rounded-full"></div>
					<span class="truncate">{user} {user === username ? '(You)' : ''}</span>
				</div>
			{/each}
		</div>
	</aside>

	<!-- Main Chat -->
	<div class="flex-1 flex flex-col min-w-0">
		<!-- Header -->
		<header class="p-4 bg-white border-b flex items-center md:hidden">
			<button
				class="p-2 -ml-2 text-gray-600 hover:text-gray-900"
				onclick={toggleSidebar}
				aria-label="Toggle sidebar"
			>
				<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
				</svg>
			</button>
			<h1 class="ml-2 font-semibold">Chat</h1>
		</header>

		<!-- Messages -->
		<div bind:this={chatContainer} class="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-50">
			{#each messages as msg (msg.id || msg.created_at)}
				<div class="flex flex-col {msg.sender === username ? 'items-end' : 'items-start'}">
					<div class="flex items-baseline space-x-2">
						<span class="text-xs font-semibold text-gray-500">{msg.sender}</span>
						<span class="text-[10px] text-gray-400">{formatTime(msg.created_at)}</span>
					</div>
					<div
						class="mt-1 px-4 py-2 rounded-lg max-w-[85%] md:max-w-md break-words {msg.sender === username
							? 'bg-blue-600 text-white'
							: 'bg-white text-gray-800 shadow-sm'}"
					>
						{msg.content}
					</div>
				</div>
			{/each}

			{#if typingUsers.size > 0}
				<div class="text-xs text-gray-400 italic px-2">
					{Array.from(typingUsers).join(', ')}
					{typingUsers.size === 1 ? 'is' : 'are'} typing...
				</div>
			{/if}
		</div>

		<!-- Input -->
		<footer class="p-4 bg-white border-t">
			<div class="flex flex-wrap gap-1 mb-2">
				{#each emojis as emoji (emoji)}
					<button onclick={() => addEmoji(emoji)} class="hover:bg-gray-100 p-1.5 rounded text-lg">
						{emoji}
					</button>
				{/each}
			</div>
			<div class="flex space-x-2">
				<input
					type="text"
					bind:value={newMessage}
					onkeydown={handleKeydown}
					placeholder="Type a message..."
					class="flex-1 border border-gray-300 rounded-full px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
				/>
				<button
					onclick={sendMessage}
					class="bg-blue-600 text-white rounded-full px-5 md:px-6 py-2 hover:bg-blue-700 focus:outline-none text-sm font-medium"
				>
					Send
				</button>
			</div>
		</footer>
	</div>
</div>
