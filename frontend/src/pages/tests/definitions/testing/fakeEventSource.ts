type Listener = (event: { data: string }) => void;

// FakeEventSource is the hub event stream in tests: emit pushes an event
export class FakeEventSource {
    private listeners = new Map<string, Set<Listener>>();

    addEventListener(name: string, listener: Listener): void {
        this.listeners.set(name, (this.listeners.get(name) ?? new Set()).add(listener));
    }

    removeEventListener(name: string, listener: Listener): void {
        this.listeners.get(name)?.delete(listener);
    }

    close(): void {
        this.listeners.clear();
    }

    emit(name: string, data: unknown): void {
        this.listeners.get(name)?.forEach(listener => listener({data: JSON.stringify(data)}));
    }
}
