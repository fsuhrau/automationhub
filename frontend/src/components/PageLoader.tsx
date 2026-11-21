import React from "react";
import {Alert, Backdrop, Box, Button, CircularProgress} from "@mui/material";
import {errorMessage} from "../lib/errorMessage";

interface PageLoaderProps {
    // an error from loading; shows the message and a retry button
    error?: unknown,
    onRetry?: () => void,
}

// PageLoader is the shared loading state of pages: a spinner while loading,
// or the error with a retry button instead of spinning forever.
const PageLoader: React.FC<PageLoaderProps> = ({error, onRetry}) => {
    if (error) {
        return <Box sx={{p: 2}}>
            <Alert severity="error" action={onRetry && <Button color="inherit" size="small" onClick={onRetry}>Retry</Button>}>
                {errorMessage(error)}
            </Alert>
        </Box>;
    }
    return <Backdrop sx={{zIndex: 1, color: '#fff'}} open={true}><CircularProgress color="inherit"/></Backdrop>;
};

export default PageLoader;
