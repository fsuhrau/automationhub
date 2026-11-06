import React from 'react';
import {Button, IconButton, Stack, Tooltip} from '@mui/material';
import {DataGrid, GridColDef} from "@mui/x-data-grid";
import IDeviceData from '../types/device';
import Grid from '@mui/system/Grid';
import {ArrowForward, PlayArrow, LockOpen, ScreenShare} from "@mui/icons-material";
import {useNavigate} from "react-router-dom";
import {useProjectContext} from "../hooks/ProjectProvider";
import DeviceStatus from "./DeviceStatus";
import DeviceController from "./DeviceController";

interface DevicesTableProps {
    devices: IDeviceData[]
    onSelectForRun: (deviceId: number) => void;
    onOpenDeviceDetails: (deviceId: number) => void;
    onUnlockDevice: (deviceId: number) => void;
}

const DevicesTable: React.FC<DevicesTableProps> = (props: DevicesTableProps) => {

    const {devices, onSelectForRun, onOpenDeviceDetails, onUnlockDevice} = props;

    const navigate = useNavigate();
    const {projectIdentifier} = useProjectContext();

    // every action is labelled; the row itself opens the device details
    const stop = (e: React.MouseEvent, action: () => void) => {
        e.stopPropagation();
        action();
    };
    const renderActions = (device: any) => {
        return <Stack direction="row" spacing={0.5} justifyContent="flex-end" alignItems="center" sx={{height: '100%'}}>
            {device.dev && (
                <Tooltip title="Live control: see the screen and use the device">
                    <IconButton size="small" color="primary" aria-label="Live control"
                                onClick={e => stop(e, () => navigate(`/project/${projectIdentifier}/device/${device.id}/control`))}>
                        <ScreenShare fontSize="small"/>
                    </IconButton>
                </Tooltip>)
            }
            {device.isLocked && (
                <Tooltip title="Unlock: end the test run or control session that holds the device">
                    <IconButton size="small" color="warning" aria-label="Unlock device"
                                onClick={e => stop(e, () => onUnlockDevice(device.id))}>
                        <LockOpen fontSize="small"/>
                    </IconButton>
                </Tooltip>)
            }
            {device.connection && (
                <Button color="primary" size="small" variant="outlined" startIcon={<PlayArrow/>}
                        onClick={e => stop(e, () => onSelectForRun(device.id))}>
                    Run
                </Button>)
            }
            <Tooltip title="Details">
                <IconButton size="small" aria-label="Device details"
                            onClick={e => stop(e, () => onOpenDeviceDetails(device.id))}>
                    <ArrowForward fontSize="small"/>
                </IconButton>
            </Tooltip>
        </Stack>
    }

    const columns: GridColDef[] = [
        {
            field: 'id',
            headerName: 'ID',
        },
        {
            field: 'name',
            headerName: 'Name/Identifier',
            flex: 1,
            minWidth: 90,
            renderCell: (params) => (<Grid container={true}>
                <Grid size={12}>{params.row.alias.length > 0 ? params.row.alias : params.row.name}</Grid>
                <Grid size={12}>{params.row.deviceIdentifier}</Grid>
            </Grid>)
        },
        {
            field: 'os',
            headerName: 'OS',
            flex: 1,
            minWidth: 90,
        },
        {
            field: 'osVersion',
            headerName: 'Version',
            flex: 0.5,
            minWidth: 90,
        },
        {
            field: 'status',
            headerName: 'Status',
            headerAlign: 'right',
            align: 'right',
            flex: 1,
            minWidth: 100,
            renderCell: (params) => (
                <Stack direction="row" spacing={0.5} alignItems="center" justifyContent="flex-end"
                       sx={{height: '100%'}}>
                    <DeviceStatus state={params.value} platform={params.row.platformType}/>
                    <DeviceController owner={params.row.lockedBy}/>
                </Stack>),
        },
        {
            field: 'actions',
            headerName: '',
            headerAlign: 'right',
            align: 'right',
            sortable: false,
            flex: 1,
            minWidth: 160,
            renderCell: (params) => renderActions(params.row),
        },
    ];

    return (
        <DataGrid
            disableRowSelectionOnClick
            onRowClick={params => onOpenDeviceDetails(params.id as number)}
            sx={{'& .MuiDataGrid-row': {cursor: 'pointer'}}}
            rows={devices}
            columns={columns}
            getRowClassName={(params) =>
                params.indexRelativeToCurrentPage % 2 === 0 ? 'even' : 'odd'
            }
            getRowId={(row) => row.id}
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
