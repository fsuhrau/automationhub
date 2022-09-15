import React from 'react';
import { createRoot } from 'react-dom/client';
import { ThemeProvider } from "@mui/material/styles";
import theme from './style/theme';
import { AppRouter } from "./router/approuter";

const container = document.getElementById('root');
const root = createRoot(container!);
root.render(<React.StrictMode>
    <ThemeProvider theme={ theme }>
       <AppRouter />
    </ThemeProvider>
</React.StrictMode>);
