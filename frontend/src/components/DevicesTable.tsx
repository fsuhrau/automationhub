import React from 'react';
import {Button, ButtonGroup,} from '@mui/material';
import {DataGrid, GridColDef} from "@mui/x-data-grid";
import IDeviceData, { deviceState } from '../types/device';
import Grid from '@mui/system/Grid';
import {ArrowForward, PlayArrow, LockOpen} from "@mui/icons-material";

interface DevicesTableProps {
    devices: IDeviceData[]
    onSelectForRun: (deviceId: number) => void;
    onOpenDeviceDetails: (deviceId: number) => void;
    onUnlockDevice: (deviceId: number) => void;
}

const DevicesTable: React.FC<DevicesTableProps> = (props: DevicesTableProps) => {

    const {devices, onSelectForRun, onOpenDeviceDetails, onUnlockDevice} = props;

    const renderActions = (device: any) => {
        return <ButtonGroup variant={"text"} aria-label="text button group">
            {device.Status === 5 && (
                <Button color="primary" size="small" variant="outlined"
                        endIcon={<LockOpen/>}
                        onClick={(e) => onUnlockDevice(device.ID)}>
                    Run
                </Button>)
            }
            {device.Connection && (
                <Button color="primary" size="small" variant="outlined"
                        endIcon={<PlayArrow/>}
                        onClick={(e) => onSelectForRun(device.ID)}>
                    Run
                </Button>)
            }
            <Button color="primary" size={'small'}
                    onClick={(e) => {
                        onOpenDeviceDetails(device.ID);
                    }}>
                <ArrowForward/>
            </Button>
        </ButtonGroup>
    }

    const columns: GridColDef[] = [
        {
            field: 'ID',
            headerName: 'ID',
        },
        {
            field: 'Name',
            headerName: 'Name/Identifier',
            flex: 1,
            minWidth: 90,
            renderCell: (params) => (<Grid container={true}>
                <Grid size={12}>{params.row.Alias.length > 0 ? params.row.Alias : params.row.Name}</Grid>
                <Grid size={12}>{params.row.DeviceIdentifier}</Grid>
            </Grid>)
        },
        {
            field: 'OS',
            headerName: 'OS',
            flex: 1,
            minWidth: 90,
        },
        {
            field: 'OSVersion',
            headerName: 'Version',
            flex: 0.5,
            minWidth: 90,
        },
        {
            field: 'Status',
            headerName: 'Status',
            headerAlign: 'right',
            align: 'right',
            flex: 1,
            minWidth: 100,
            renderCell: (params) => deviceState(params.value),
        },
        {
            field: 'Session',
            headerName: 'Session',
            headerAlign: 'right',
            align: 'right',
            flex: 0.5,
            minWidth: 90,
        },
        {
            field: 'actions',
            headerName: '',
            headerAlign: 'right',
            align: 'right',
            flex: 1,
            minWidth: 100,
            renderCell: (params) => renderActions(params.row),
        },
    ];

    return (
        <DataGrid
            disableRowSelectionOnClick
            rows={devices}
            columns={columns}
            getRowClassName={(params) =>
                params.indexRelativeToCurrentPage % 2 === 0 ? 'even' : 'odd'
            }
            getRowId={(row) => row.ID}
            initialState={{
                pagination: {paginationModel: {pageSize: 20}},
            }}
            pageSizeOptions={[10, 20, 50]}
            disableColumnResize
            density="compact"
            slotProps={{
                filterPanel: {
                    filterFormProps: {
                        logicOperatorInputProps: {
                            variant: 'outlined',
                            size: 'small',
                        },
                        columnInputProps: {
                            variant: 'outlined',
                            size: 'small',
                            sx: {mt: 'auto'},
                        },
                        operatorInputProps: {
                            variant: 'outlined',
                            size: 'small',
                            sx: {mt: 'auto'},
                        },
                        valueInputProps: {
                            InputComponentProps: {
                                variant: 'outlined',
                                size: 'small',
                            },
                        },
                    },
                },
            }}
        />
    );
};

export default DevicesTable;
