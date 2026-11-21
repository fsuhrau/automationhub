import React from 'react';
import {Box} from '@mui/material';
import {iconUrlOf, tagsText} from './bundlesModel';
import {BundleDialogs} from './parts/BundleDialogs';
import {BundlesContent} from './parts/BundlesContent';
import {BundlesHeader} from './parts/BundlesHeader';
import {DropOverlay} from './parts/DropOverlay';
import {PlatformNotice} from './parts/PlatformNotice';
import {RowMenu} from './parts/RowMenu';
import {UploadQueue} from './parts/UploadQueue';
import {useBundlesPage} from './useBundlesPage';

// AppBundlesPage lists the builds of the app picked in the menu and takes
// new ones, by button or by dropping files anywhere on the page.
const AppBundlesPage: React.FC = () => {
    const page = useBundlesPage();
    const {scope, accept, drop, loader, uploads, dialog, setDialog, setMenu} = page;
    const {app, ref, routes} = scope;
    return (
        <Box {...drop.getRootProps()} sx={{width: '100%', maxWidth: {sm: '100%', md: '1700px'}, mt: 2, position: 'relative', outline: 'none'}}>
            <input {...drop.getInputProps()} aria-label="Bundle files"/>
            {drop.isDragActive && <DropOverlay appName={app?.name} accept={accept.label}/>}
            <BundlesHeader app={app} accept={accept.label} onUpload={drop.open}
                compareUrl={(loader.data?.length ?? 0) >= 2 ? routes.compareBundles() : undefined}/>
            <PlatformNotice platform={app?.platform}/>
            <UploadQueue uploads={uploads}/>
            <BundlesContent loader={loader} accept={accept.label} onChoose={drop.open} ctx={{
                fresh: page.fresh, now: page.now, detailUrl: routes.bundle, iconUrl: b => iconUrlOf(ref, b),
                onTags: b => setDialog({kind: 'tags', bundle: b, text: tagsText(b)}),
                onMenu: (anchor, bundle) => setMenu({anchor, bundle}),
            }}/>
            <RowMenu target={page.menu} bundles={loader.data ?? []} routes={routes} onDialog={setDialog} onClose={() => setMenu(null)}/>
            <BundleDialogs dialog={dialog} actions={page.actions} onChange={setDialog}/>
        </Box>
    );
};

export default AppBundlesPage;
