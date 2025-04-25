import * as React from 'react';
import {DataGrid, GridColDef} from '@mui/x-data-grid';
import Chip from "@mui/material/Chip";
import {TestResultState} from "../types/test.result.state.enum";
import {getTestStatusChipColor, getTestStatusText} from "../helper/TestStatusHelper";

function renderStatus(status: TestResultState) {
    return <Chip label={getTestStatusText(status)} color={getTestStatusChipColor(status)} size="small"/>;
}

export const columns: GridColDef[] = [
    {
        field: 'name',
        headerName: 'test name',
        flex: 1.5,
        minWidth: 300
        /*
                                                    <Link
                                                href={`/project/${projectId}/app/${data.testRun.test.appId}/test/4/run/${data.testRunId}/${data.id}`}
                                                underline="none">
                                                {testName.length > 1 ? testName[1] : data.testName}
                                            </Link> <br/>
                                            {data.device && (data.device.alias.length > 0 ? data.device.alias : data.device.name)}

         */
    },
    {
        field: 'result',
        headerName: 'Result',
        flex: 0.5,
        minWidth: 90,
        renderCell: (params) => renderStatus(params.value as any),
    },
    /*
    {
        field: 'fps',
        headerName: 'Median fps',
        headerAlign: 'right',
        align: 'right',
        flex: 1,
        minWidth: 80,
    },
    {
        field: 'mem',
        headerName: 'Median Memory',
        headerAlign: 'right',
        align: 'right',
        flex: 1,
        minWidth: 100,
    },
    {
        field: 'cpu',
        headerName: 'Median cpu',
        headerAlign: 'right',
        align: 'right',
        flex: 1,
        minWidth: 120,
    },
     */
    {
        field: 'time',
        headerName: 'Time',
        headerAlign: 'right',
        align: 'right',
        flex: 1,
        minWidth: 100,
    },
    /*
  {
    field: 'conversions',
    headerName: 'Daily Conversions',
    flex: 1,
    minWidth: 150,
    renderCell: renderSparklineCell,
  },
     */
];

export interface ProjectDashboardTestResultsData {
    id: number,
    name: string,
    result: TestResultState,
    fps: string,
    mem: string,
    cpu: string,
    time: string,
}

export type ProjectDashboardTestResultsDataGridProps = {
    data: ProjectDashboardTestResultsData[];
};

export default function ProjectDashboardTestResultsDataGrid(props: ProjectDashboardTestResultsDataGridProps) {
    const {data} = props;
    return (
        <DataGrid
            autoHeight
            rows={data}
            columns={columns}
            getRowClassName={(params) =>
                params.indexRelativeToCurrentPage % 2 === 0 ? 'even' : 'odd'
            }
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
}
