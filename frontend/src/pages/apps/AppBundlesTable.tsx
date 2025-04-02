import React, {useEffect, useState} from 'react';
import {Button, ButtonGroup, Typography,} from '@mui/material';
import {DataGrid, GridColDef} from "@mui/x-data-grid";
import {useProjectContext} from '../../hooks/ProjectProvider';
import {useError} from "../../ErrorProvider";
import {IAppBinaryData, prettySize} from "../../types/app";
import {deleteAppBundle, getAppBundles} from "../../services/app.service";
import DownloadIcon from "@mui/icons-material/Download";
import DeleteForeverIcon from "@mui/icons-material/DeleteForever";
import Moment from "react-moment";

interface AppBundlesTableProps {
    appId: number | null
}

const AppBundlesTable: React.FC<AppBundlesTableProps> = (props: AppBundlesTableProps) => {

    const {appId} = props;
    const {project, projectIdentifier} = useProjectContext();
    const {setError} = useError()

    const app = project.Apps.find(a => a.ID === appId);
    const [bundles, setBundles] = useState<IAppBinaryData[]>([]);

    const renderActions = (bundle: any) => {
        return <ButtonGroup variant={"text"} aria-label="text button group">
            <Button size="small" startIcon={<DownloadIcon/>} href={`/upload/${bundle?.AppPath}`}></Button>
            <Button size="small" startIcon={<DeleteForeverIcon/>} onClick={() => {
                handleDeleteApp(bundle?.ID as number);
            }}></Button>
        </ButtonGroup>
    }

    const columns: GridColDef[] = [
        {
            field: 'ID',
            headerName: 'ID',
            flex: 0.5,
            minWidth: 1
        },
        {
            field: 'Version',
            headerName: 'Version',
            flex: 0.5,
            minWidth: 90,
        },
        {
            field: 'CreatedAt',
            headerName: 'Created',
            flex: 1,
            minWidth: 90,
            renderCell: (params) => (<Moment
                format="YYYY/MM/DD HH:mm:ss">{params.value}</Moment>)
        },
        {
            field: 'Name',
            headerName: 'Name',
            flex: 0.5,
            minWidth: 90,
        },
        {
            field: 'Size',
            headerName: 'Size',
            headerAlign: 'right',
            align: 'right',
            flex: 1,
            minWidth: 100,
            renderCell: (params) => (<Typography variant={'caption'}>{prettySize(params.value)}</Typography>)
        },
        {
            field: 'Tags',
            headerName: 'Tags',
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

    useEffect(() => {
        if (projectIdentifier !== null && appId != null) {
            if (bundles.length == 0) {
                getAppBundles(projectIdentifier, appId as number).then(response => {
                    setBundles(response.data);
                }).catch(ex => {
                    setError(ex);
                });
            }
        }
    }, [projectIdentifier, appId]);

    const handleDeleteApp = (bundleId: number): void => {
        deleteAppBundle(projectIdentifier, app?.ID as number, bundleId).then(value => {
            setBundles(prevState => {
                const newState = [...prevState];
                const index = newState.findIndex(value1 => value1.ID == bundleId);
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

export default AppBundlesTable;
