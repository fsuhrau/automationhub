import React from "react";
import {BrowserRouter} from "react-router-dom";
import {AuthProvider} from "./hooks/AuthProvider";
import AppTheme from "./shared-theme/AppTheme";
import {
    chartsCustomizations,
    dataGridCustomizations,
    datePickersCustomizations,
    treeViewCustomizations
} from "./theme/customizations";
import {AppProviders} from "./app/providers";
import {AppRoutes} from "./app/routes";

const xThemeComponents = {
    ...chartsCustomizations,
    ...dataGridCustomizations,
    ...datePickersCustomizations,
    ...treeViewCustomizations,
};

declare module '@mui/material/Paper' {
    interface PaperPropsVariantOverrides {
        paper_content: true;
    }
}

type AppProps = Omit<React.ComponentProps<typeof AppTheme>, 'children' | 'themeComponents'>;

const App: React.FC<AppProps> = props => (
    <div className="App">
        <AppTheme {...props} themeComponents={xThemeComponents}>
            <AppProviders>
                <BrowserRouter>
                    <AuthProvider>
                        <AppRoutes/>
                    </AuthProvider>
                </BrowserRouter>
            </AppProviders>
        </AppTheme>
    </div>
);

export default App;
