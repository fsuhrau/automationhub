import React, {useEffect, useState} from 'react';
import {IconButton} from '@mui/material';
import {DataGrid, GridColDef} from "@mui/x-data-grid";
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useError} from "../../ErrorProvider";
import {IAppBinaryData} from "../../types/app";
import {deleteAppBundle, getAppBundles} from "../../services/app.service";
import DownloadIcon from "@mui/icons-material/Download";
import DeleteForeverIcon from "@mui/icons-material/DeleteForever";
import DateTime from "../../components/DateTime";
import {byteFormat} from "../tests/value_formatter";
import {useHubState} from "../../hooks/HubStateProvider";

interface AppBundlesTableProps {
    appId: number | null
}

const AppBundlesTable: React.FC<AppBundlesTableProps> = (props: AppBundlesTableProps) => {

    const {appId} = props;
    const {project, projectIdentifier} = useProjectContext();
    const {state} = useHubState();
    const {setError} = useError()

    const app = state.apps?.find(a => a.id === appId);
    const [bundles, setBundles] = useState<IAppBinaryData[]>([]);

    const renderActions = (bundle: any) => {
        return <>
            <IconButton size="small" title="Download" href={`/upload/${bundle?.appPath}`}>
                <DownloadIcon fontSize="small"/>
            </IconButton>
            <IconButton size="small" title="Delete" onClick={() => handleDeleteApp(bundle?.id as number)}>
                <DeleteForeverIcon fontSize="small"/>
            </IconButton>
        </>
    }

    // fixed widths for short values; the name takes the remaining space
    const columns: GridColDef[] = [
        {
            field: 'id',
            headerName: 'ID',
            width: 70,
        },
        {
            field: 'name',
            headerName: 'Name',
            flex: 2,
            minWidth: 200,
            renderCell: (params) => <span title={params.value}>{params.value}</span>,
        },
        {
            field: 'version',
            headerName: 'Version',
            flex: 2,
            minWidth: 200,
            renderCell: (params) => <span title={params.value}>{params.value}</span>,
        },
        {
            field: 'createdAt',
            headerName: 'Created',
            width: 170,
            renderCell: (params) => (<DateTime format="YYYY/MM/DD HH:mm:ss" value={params.value}/>)
        },
        {
            field: 'size',
            headerName: 'Size',
            headerAlign: 'right',
            align: 'right',
            width: 100,
            valueFormatter: (value) => byteFormat(value as number),
        },
        {
            field: 'tags',
            headerName: 'Tags',
            flex: 0.5,
            minWidth: 80,
        },
        {
            field: 'actions',
            headerName: '',
            headerAlign: 'right',
            align: 'right',
            width: 96,
            sortable: false,
            renderCell: (params) => renderActions(params.row),
        },
    ];

    useEffect(() => {
        if (projectIdentifier !== null && appId != null) {
            getAppBundles(projectIdentifier, appId as number).then(appBundles => {
                setBundles(appBundles);
            }).catch(ex => {
                setError(ex);
            });
        }
    }, [projectIdentifier, appId]);

    const handleDeleteApp = (bundleId: number): void => {
        deleteAppBundle(projectIdentifier, app?.id as number, bundleId).then(value => {
            setBundles(prevState => {
                const newState = [...prevState];
                const index = newState.findIndex(value1 => value1.id == bundleId);
                if (index > -1) {
                    newState.splice(index, 1);
                }
                return newState;
            });
        }).catch(ex => setError(ex));
    };


    return (
        <DataGrid
            disableRowSelectionOnClick
            rows={bundles}
            columns={columns}
            getRowClassName={(params) =>
                params.indexRelativeToCurrentPage % 2 === 0 ? 'even' : 'odd'
            }
            getRowId={(row) => row.id}
            initialState={{
                pagination: {paginationModel: {pageSize: 20}},
            }}
            pageSizeOptions={[10, 20, 50]}
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

export default AppBundlesTable;
