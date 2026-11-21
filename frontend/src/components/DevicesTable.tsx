import React from 'react';
import {DataGrid} from '@mui/x-data-grid';
import {useNavigate} from 'react-router-dom';
import IDeviceData from '../types/device';
import {useProjectContext} from '../hooks/ProjectProvider';
import {routes} from '../lib/routes';
import {deviceColumns} from '../pages/devices/parts/deviceColumns';

interface DevicesTableProps {
    devices: IDeviceData[]
    onSelectForRun: (deviceId: number) => void;
    onOpenDeviceDetails: (deviceId: number) => void;
    onUnlockDevice: (deviceId: number) => void;
}

const inputProps = {variant: 'outlined', size: 'small'} as const;

const filterPanel = {
    filterFormProps: {
        logicOperatorInputProps: inputProps,
        columnInputProps: {...inputProps, sx: {mt: 'auto'}},
        operatorInputProps: {...inputProps, sx: {mt: 'auto'}},
        valueInputProps: {InputComponentProps: inputProps},
    },
};

const DevicesTable: React.FC<DevicesTableProps> = ({devices, ...handlers}) => {
    const navigate = useNavigate();
    const {projectIdentifier} = useProjectContext();
    const openControl = (id: number): void => {
        void navigate(routes.project(projectIdentifier).device(id).control());
    };
    return (
        <DataGrid
            disableRowSelectionOnClick={true}
            onRowClick={params => handlers.onOpenDeviceDetails(params.id as number)}
            sx={{'& .MuiDataGrid-row': {cursor: 'pointer'}}}
            rows={devices}
            columns={deviceColumns({...handlers, onOpenControl: openControl})}
            getRowClassName={(params) => params.indexRelativeToCurrentPage % 2 === 0 ? 'even' : 'odd'}
            getRowId={(row) => row.id}
            initialState={{pagination: {paginationModel: {pageSize: 20}}}}
            pageSizeOptions={[10, 20, 50]}
            disableColumnResize={true}
            density="compact"
            slotProps={{filterPanel}}
        />
    );
};

export default DevicesTable;
