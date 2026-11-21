import {expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {createTheme, ThemeProvider} from '@mui/material/styles';
import AppearanceSection from './AppearanceSection';

const theme = createTheme({colorSchemes: {light: true, dark: true}, defaultColorScheme: 'light'});

it('AppearanceSection: shows the color mode and switches it', async () => {
    render(<ThemeProvider theme={theme}><AppearanceSection/></ThemeProvider>);
    const group = await screen.findByRole('group', {name: 'Color mode'});
    expect(group).toBeInTheDocument();
    await userEvent.setup().click(screen.getByRole('button', {name: 'Dark'}));
    expect(screen.getByRole('button', {name: 'Dark'})).toHaveAttribute('aria-pressed', 'true');
});

it('AppearanceSection: offers nothing without a theme that has color schemes', () => {
    render(<AppearanceSection/>);
    expect(screen.queryByRole('group', {name: 'Color mode'})).toBeNull();
});
