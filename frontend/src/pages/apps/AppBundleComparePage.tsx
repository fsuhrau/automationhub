import React from 'react';
import {Alert, Box, LinearProgress} from '@mui/material';
import {useNow} from '../../hooks/useNow';
import {CompareHeading} from './parts/CompareHeading';
import {ComparePickerCard} from './parts/ComparePickerCard';
import {CompareResult} from './parts/CompareResult';
import {useBundleCompare} from './useBundleCompare';
import {useBundleScope} from './useBundleScope';

const MINUTE = 60_000;

// AppBundleComparePage puts two bundles of the app side by side and shows
// what changed from the first to the second: attributes, lists like the
// permissions, and where the size went. Archived bundles compare too.
const AppBundleComparePage: React.FC = () => {
    const scope = useBundleScope();
    const compare = useBundleCompare(scope.ref);
    const now = useNow(MINUTE);
    const {from, to, error, loading} = compare;
    return (
        <Box sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2}}>
            <CompareHeading appName={scope.app?.name} back={scope.routes.bundles()}/>
            {error && <Alert severity="error" sx={{mb: 2}}>{error}</Alert>}
            <ComparePickerCard compare={compare} now={now}/>
            {loading && <LinearProgress/>}
            {from && to && <CompareResult scope={scope} from={from} to={to} now={now}/>}
        </Box>
    );
};

export default AppBundleComparePage;
