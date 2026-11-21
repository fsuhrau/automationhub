import React from 'react';
import {Chip, Stack} from '@mui/material';
import {getPlatformName, PlatformType} from '../../../types/platform.type.enum';
import PlatformTypeIcon from '../../../components/PlatformTypeIcon';
import {PlatformChoice, PlatformCounts} from '../deviceListModel';

interface PlatformFilterProps {
    total: number,
    counts: PlatformCounts,
    platforms: PlatformType[],
    selected: PlatformChoice,
    onSelect: (platform: PlatformChoice) => void,
}

interface ChoiceProps {
    label: string,
    active: boolean,
    icon?: React.ReactElement,
    onClick: () => void,
}

const Choice: React.FC<ChoiceProps> = ({label, active, icon, onClick}) =>
    <Chip size="small" label={label} icon={icon} color={active ? 'primary' : 'default'}
        variant={active ? 'filled' : 'outlined'} onClick={onClick}/>;

// PlatformFilter narrows the device pool to one platform; it only shows
// when the pool has more than one
export const PlatformFilter: React.FC<PlatformFilterProps> = ({total, counts, platforms, selected, onSelect}) => {
    if (platforms.length < 2) return null;
    return <Stack direction="row" alignItems="center" spacing={1} useFlexGap={true} sx={{flexWrap: 'wrap'}}
        role="group" aria-label="Filter by platform">
        <Choice label={`All (${total})`} active={selected === 'all'} onClick={() => onSelect('all')}/>
        {platforms.map(p => <Choice key={p} label={`${getPlatformName(p)} (${counts[p]})`} active={selected === p}
            icon={<PlatformTypeIcon platformType={p} fontSize="small"/>} onClick={() => onSelect(p)}/>)}
    </Stack>;
};
