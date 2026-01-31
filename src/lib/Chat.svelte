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
</script>

<div class="flex h-screen bg-gray-100">
	<!-- Sidebar -->
	<div class="w-64 bg-white border-r flex flex-col">
		<div class="p-4 border-b font-bold text-lg">Online Users ({onlineUsers.length})</div>
		<div class="flex-1 overflow-y-auto p-4 space-y-2">
			{#each onlineUsers as user (user)}
				<div class="flex items-center space-x-2">
					<div class="w-2 h-2 bg-green-500 rounded-full"></div>
					<span>{user} {user === username ? '(You)' : ''}</span>
				</div>
			{/each}
		</div>
	</div>

	<!-- Main Chat -->
	<div class="flex-1 flex flex-col">
		<!-- Messages -->
		<div bind:this={chatContainer} class="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-50">
			{#each messages as msg (msg.id || msg.created_at)}
				<div class="flex flex-col {msg.sender === username ? 'items-end' : 'items-start'}">
					<div class="flex items-baseline space-x-2">
						<span class="text-xs font-semibold text-gray-500">{msg.sender}</span>
						<span class="text-[10px] text-gray-400">{formatTime(msg.created_at)}</span>
					</div>
					<div
						class="mt-1 px-4 py-2 rounded-lg max-w-md {msg.sender === username
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
		<div class="p-4 bg-white border-t">
			<div class="flex flex-wrap gap-2 mb-2">
				{#each emojis as emoji (emoji)}
					<button onclick={() => addEmoji(emoji)} class="hover:bg-gray-100 p-1 rounded">
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
					class="flex-1 border border-gray-300 rounded-full px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
				/>
				<button
					onclick={sendMessage}
					class="bg-blue-600 text-white rounded-full px-6 py-2 hover:bg-blue-700 focus:outline-none"
				>
					Send
				</button>
			</div>
		</div>
	</div>
</div>
