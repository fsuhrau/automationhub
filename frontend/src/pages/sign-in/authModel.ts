// The sign in and sign up forms without React: the rules of their fields.

export const isEmail = (value: string): boolean => /\S+@\S+\.\S+/.test(value);

// the master's rule for new passwords
export const MIN_PASSWORD = 8;

// the page that asked for the sign in, which it returns to afterwards
export const returnPath = (state: unknown, fallback: string): string => (state as { from?: string } | null)?.from ?? fallback;

export interface SignInFields {
    email: string,
    password: string,
}

export interface SignUpFields extends SignInFields {
    name: string,
}

export type FieldErrors<T> = Partial<Record<keyof T, string>>;

export const signInErrors = ({email, password}: SignInFields): FieldErrors<SignInFields> => ({
    email: isEmail(email) ? undefined : 'Enter the email address of your account.',
    password: password ? undefined : 'Enter your password.',
});

export const signUpErrors = ({name, email, password}: SignUpFields): FieldErrors<SignUpFields> => ({
    name: name.trim() ? undefined : 'Enter your name; others see it on devices you control.',
    email: isEmail(email) ? undefined : 'Enter an email address; you sign in with it.',
    password: password.length >= MIN_PASSWORD ? undefined : `At least ${MIN_PASSWORD} characters.`,
});

export const hasErrors = <T, >(errors: FieldErrors<T>): boolean => Object.values(errors).some(Boolean);
