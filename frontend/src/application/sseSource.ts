// Event that is dispatched on window whenever the hub event stream
// (re)connects. EventSource reconnects on its own after the master restarts,
// but events published while it was down are lost, so listeners reload state.
export const SSE_CONNECTED_EVENT = 'hub-sse-connected';

export const createHubEventSource = (endpoint: string) => (): EventSource => {
    const source = new EventSource(endpoint);
    source.addEventListener('open', () => window.dispatchEvent(new Event(SSE_CONNECTED_EVENT)));
    return source;
};
