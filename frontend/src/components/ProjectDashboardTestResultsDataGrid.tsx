import * as React from 'react';
import {DataGrid, GridColDef, GridRowsProp} from '@mui/x-data-grid';
import ITestProtocolData from "../types/test.protocol";
import Chip from "@mui/material/Chip";
import {TestResultState} from "../types/test.result.state.enum";
import {getTestStatusChipColor, getTestStatusText} from "../helper/TestStatusHelper";

function renderStatus(status: TestResultState) {
    return <Chip label={getTestStatusText(status)} color={getTestStatusChipColor(status)} size="small"/>;
}

export const columns: GridColDef[] = [
    {
        field: 'name',
        headerName: 'Test Name',
        flex: 1.5,
        minWidth: 300
        /*
                                                    <Link
                                                href={`/project/${projectId}/app/${data.TestRun.Test.AppID}/test/4/run/${data.TestRunID}/${data.ID}`}
                                                underline="none">
                                                {testName.length > 1 ? testName[1] : data.TestName}
                                            </Link> <br/>
                                            {data.Device && (data.Device.Alias.length > 0 ? data.Device.Alias : data.Device.Name)}

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
        headerName: 'Median FPS',
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
        headerName: 'Median CPU',
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

export type ProjectDashboardTestResultsDataGridProps = {
    data: GridRowsProp[];
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
