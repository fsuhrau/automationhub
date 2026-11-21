import React, {useMemo, useState} from 'react';
import {Stack} from '@mui/material';
import {IAppBundleDetail} from '../../../types/app';
import {BundleScope} from '../useBundleScope';
import {compareBundles, hideEmpty, onlyChanges} from '../bundleCompareModel';
import {iconUrlOf} from '../bundlesModel';
import {CompareBundleCard} from './CompareBundleCard';
import {CompareSectionCard} from './CompareSectionCard';
import {CompareSummary} from './CompareSummary';
import {SizesCard} from './SizesCard';

interface CompareResultProps {
    scope: BundleScope,
    from: IAppBundleDetail,
    to: IAppBundleDetail,
    now: number,
}

const PARTS_NOTE = 'Installed size of each part of the bundle';
const FILES_NOTE = 'The largest files of either bundle; a file missing on one side may just be smaller there';

// CompareResult is everything below the pickers: the two bundles, how much
// changed and the changes themselves
export const CompareResult: React.FC<CompareResultProps> = ({scope, from, to, now}) => {
    const [changesOnly, setChangesOnly] = useState(true);
    const comparison = useMemo(() => compareBundles(from, to), [from, to]);
    const sections = changesOnly ? onlyChanges(comparison.sections) : hideEmpty(comparison.sections);
    const card = (side: string, bundle: IAppBundleDetail): React.ReactNode =>
        <CompareBundleCard side={side} bundle={bundle} iconUrl={iconUrlOf(scope.ref, bundle)} now={now}
            to={scope.routes.bundle(bundle.id)} platform={scope.app?.platform}/>;
    return <>
        <Stack direction={{xs: 'column', md: 'row'}} spacing={2} sx={{mb: 2}}>
            {card('From', from)}
            {card('To', to)}
        </Stack>
        <CompareSummary from={from} to={to} changes={comparison.changes} changesOnly={changesOnly} onChangesOnly={setChangesOnly}/>
        <Stack spacing={2}>
            {sections.map(s => <CompareSectionCard key={s.title} section={s}/>)}
            <SizesCard title="Size by part" subtitle={PARTS_NOTE} rows={comparison.parts} onlyChanged={changesOnly} missing="not there"/>
            <SizesCard title="Largest files" subtitle={FILES_NOTE} rows={comparison.files} onlyChanged={changesOnly} missing="not among largest"/>
        </Stack>
    </>;
};
