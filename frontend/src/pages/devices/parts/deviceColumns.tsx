import {Stack} from '@mui/material';
import Grid from '@mui/system/Grid';
import {GridColDef} from '@mui/x-data-grid';
import IDeviceData from '../../../types/device';
import DeviceStatus from '../../../components/DeviceStatus';
import DeviceController from '../../../components/DeviceController';
import {DeviceRowActions, RowHandlers} from './DeviceRowActions';

const nameColumn: GridColDef<IDeviceData> = {
    field: 'name',
    headerName: 'Name/Identifier',
    flex: 1,
    minWidth: 90,
    renderCell: ({row}) => <Grid container={true}>
        <Grid size={12}>{row.alias || row.name}</Grid>
        <Grid size={12}>{row.deviceIdentifier}</Grid>
    </Grid>,
};

const statusColumn: GridColDef<IDeviceData> = {
    field: 'status',
    headerName: 'Status',
    headerAlign: 'right',
    align: 'right',
    flex: 1,
    minWidth: 100,
    renderCell: ({value, row}) =>
        <Stack direction="row" spacing={0.5} alignItems="center" justifyContent="flex-end" sx={{height: '100%'}}>
            <DeviceStatus state={value} platform={row.platformType}/>
            <DeviceController owner={row.lockedBy}/>
        </Stack>,
};

const actionsColumn = (handlers: RowHandlers): GridColDef<IDeviceData> => ({
    field: 'actions',
    headerName: '',
    headerAlign: 'right',
    align: 'right',
    sortable: false,
    flex: 1,
    minWidth: 160,
    renderCell: ({row}) => <DeviceRowActions device={row} {...handlers}/>,
});

// the columns of the device table
export const deviceColumns = (handlers: RowHandlers): GridColDef<IDeviceData>[] => [
    {field: 'id', headerName: 'ID'},
    nameColumn,
    {field: 'os', headerName: 'OS', flex: 1, minWidth: 90},
    {field: 'osVersion', headerName: 'Version', flex: 0.5, minWidth: 90},
    statusColumn,
    actionsColumn(handlers),
];
