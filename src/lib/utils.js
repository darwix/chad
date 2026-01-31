export function getUrls(text) {
	const urlRegex = /https?:\/\/[^\s]+/g;
	return text.match(urlRegex) || [];
}

export function formatTime(ts) {
	return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}
