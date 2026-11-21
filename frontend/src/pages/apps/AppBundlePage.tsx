import React from 'react';
import {LinearProgress} from '@mui/material';
import {useNavigate, useParams} from 'react-router-dom';
import {useNow} from '../../hooks/useNow';
import {errorMessage} from '../../lib/errorMessage';
import {BundleDetailView} from './parts/BundleDetailView';
import {LoadFailure} from './parts/LoadFailure';
import {useBundleDetail} from './useBundleDetail';
import {useBundleDevices} from './useBundleDevices';
import {useBundleScope} from './useBundleScope';
import {useDetailActions} from './useDetailActions';

const MINUTE = 60_000;

// AppBundlePage shows everything known about one bundle: what the file says
// about the app, how it is signed, where its size goes, which devices can
// install it and which test runs used it.
const AppBundlePage: React.FC = () => {
    const scope = useBundleScope();
    const navigate = useNavigate();
    const id = Number(useParams<{ bundleId: string }>().bundleId);
    const now = useNow(MINUTE);
    const {data: detail, error, retry} = useBundleDetail(scope.ref, id);
    const devices = useBundleDevices(scope.projectId, scope.app?.platform);
    const actions = useDetailActions(scope.ref && {...scope.ref, id}, {
        onChanged: retry, onDeleted: () => navigate(scope.routes.bundles()),
    });
    if (error) return <LoadFailure message={errorMessage(error)} onRetry={retry} back={scope.routes.bundles()}/>;
    if (!detail) return <LinearProgress sx={{mt: 2}}/>;
    return <BundleDetailView detail={detail} scope={scope} devices={devices} actions={actions} now={now}/>;
};

export default AppBundlePage;
