// errorMessage turns anything that was thrown into text for the user.
// It prefers the message the server sent (axios errors, also when the body
// is unparsed JSON text, as for text requests), then the error's own
// message, then the value itself.

interface ErrorLike {
    response?: { data?: unknown },
    message?: unknown,
}

const parseJson = (text: string): unknown => {
    try {
        return JSON.parse(text);
    } catch {
        return text;
    }
};

const serverMessage = (error: ErrorLike): string | undefined => {
    const raw = error.response?.data;
    const data = typeof raw === 'string' ? parseJson(raw) : raw;
    const message = (data as { message?: unknown } | null | undefined)?.message;
    return typeof message === 'string' && message ? message : undefined;
};

const ownMessage = (error: ErrorLike): string | undefined =>
    typeof error.message === 'string' && error.message ? error.message : undefined;

export const errorMessage = (e: unknown): string => {
    if (typeof e === 'string') return e;
    if (e === null || typeof e !== 'object') return String(e);
    return serverMessage(e) ?? ownMessage(e) ?? String(e);
};
