import React, { useEffect, useState } from 'react';
import { Box, Button, Card, CardMedia, Chip, Popover } from '@mui/material';
import IProtocolEntryData from '../types/protocol.entry';
import { DataGrid, GridCellValue, GridColDef } from '@mui/x-data-grid';
import { makeStyles } from '@mui/styles';

const useStyles = makeStyles(theme => ({
    chip: {
        '& .chip--app': {
            backgroundColor: '#177E89',
            color: '#ffffff',
            margin: '5px',
        },
        '& .chip--app--unchecked': {
            backgroundColor: '#ffffff',
            fontcolor: '#177E89',
            margin: '5px',
        },
        '& .chip--testrunner': {
            backgroundColor: '#084C61',
            color: '#ffffff',
            margin: '5px',
        },
        '& .chip--testrunner--unchecked': {
            backgroundColor: '#ffffff',
            fontcolor: '#084C61',
            margin: '5px',
        },
        '& .chip--step': {
            backgroundColor: '#DB3A34',
            color: '#ffffff',
            margin: '5px',
        },
        '& .chip--step--unchecked': {
            backgroundColor: '#ffffff',
            fontcolor: '#DB3A34',
            margin: '5px',
        },
        '& .chip--status': {
            backgroundColor: '#FFC857',
            color: '#000000',
            margin: '5px',
        },
        '& .chip--status--unchecked': {
            backgroundColor: '#ffffff',
            fontcolor: '#FFC857',
            margin: '5px',
        },
        '& .chip--device': {
            backgroundColor: '#323031',
            color: '#ffffff',
            margin: '5px',
        },
        '& .chip--device--unchecked': {
            backgroundColor: '#ffffff',
            fontcolor: '#323031',
            margin: '5px',
        },
        '& .chip--action': {
            backgroundColor: '#323031',
            color: '#ffffff',
            margin: '5px',
        },
        '& .chip--action--unchecked': {
            backgroundColor: '#ffffff',
            fontcolor: '#323031',
            margin: '5px',
        },
    },
}));

interface TestProtocolContentProps {
    entries: IProtocolEntryData[]
}

const ProtocolLogComponent: React.FC<TestProtocolContentProps> = (props: TestProtocolContentProps) => {
    const classes = useStyles();
    const { entries } = props;

    const [anchorLogScreenEl, setAnchorLogScreenEl] = useState<HTMLButtonElement | null>(null);

    const showLogScreenPopup = (event: React.MouseEvent<HTMLButtonElement>): void => {
        setAnchorLogScreenEl(event.currentTarget);
    };

    const hideLogScreenPopup = (): void => {
        setAnchorLogScreenEl(null);
    };

    const logScreenOpen = Boolean(anchorLogScreenEl);
    const logScreenID = logScreenOpen ? 'simple-popover' : undefined;

    const timeFrom = (value: GridCellValue): string => {
        return new Date((value as number) * 1000).toISOString().substr(11, 8);
    };

    const nanosFrom = (value: GridCellValue): string => {
        const str = (value as number).toFixed(4);
        return str.substring(str.length - 4);
    };

    const columns: GridColDef[] = [
        {
            field: 'ID',
            headerName: 'ID',
            hide: true,
        },
        {
            field: 'Runtime',
            headerName: 'Time',
            width: 100,
            sortable: true,
            filterable: false,
            disableColumnMenu: true,
            renderCell: (params) => {
                return (<div>{ timeFrom(params.value) }.{ nanosFrom(params.value) }</div>);
            },
        },
        {
            field: 'Source',
            headerName: 'Source',
            width: 100,
            sortable: false,
            filterable: false,
            disableColumnMenu: true,
            renderCell: (params) => {
                return (<Chip className={ `chip--${ params.value }` }
                    label={ params.value }/>);
            },
        },
        {
            field: 'Level',
            headerName: 'Level',
            width: 80,
            sortable: false,
            filterable: false,
            disableColumnMenu: true,
        },
        {
            field: 'Message',
            headerName: 'Message',
            flex: 1,
            sortable: false,
            filterable: false,
            disableColumnMenu: true,
            renderCell: (params) => {
                if (params.value === '') {
                    return (
                        <div>
                            <Button aria-describedby={ params.row.ID } variant="contained"
                                onClick={ showLogScreenPopup }>
                                Show
                            </Button>
                            <Popover
                                id={ logScreenID }
                                open={ logScreenOpen }
                                anchorEl={ anchorLogScreenEl }
                                onClose={ hideLogScreenPopup }
                                anchorOrigin={ {
                                    vertical: 'bottom',
                                    horizontal: 'left',
                                } }
                            >
                                <Card>
                                    <CardMedia
                                        component="img"
                                        height="400"
                                        image={ `/api/data/${ params.row.Data }` }
                                        alt="green iguana"
                                    />
                                </Card>
                            </Popover>
                        </div>
                    );
                }
                return (<div>{ params.value }</div>);
            },
        },
    ];

    const [filteredEntries, setFilteredEntries] = useState<IProtocolEntryData[]>([]);

    const [filterApp, setFilterApp] = useState<boolean>(true);
    const [filterAction, setFilterAction] = useState<boolean>(true);
    const [filterDevice, setFilterDevice] = useState<boolean>(true);
    const [filterStatus, setFilterStatus] = useState<boolean>(true);
    const [filterStep, setFilterStep] = useState<boolean>(true);
    const [filterTestrunner, setFilterTestrunner] = useState<boolean>(true);

    const isVisible = (source: string): boolean => {
        return (filterApp && source === 'app') ||
            (filterAction && source === 'action') ||
            (filterDevice && source === 'device') ||
            (filterStatus && source === 'status') ||
            (filterStep && source === 'step') ||
            (filterTestrunner && source === 'testrunner' ||
                source === 'screen');
    };

    const filterEntries = (ents: IProtocolEntryData[]): IProtocolEntryData[] => {
        return ents.filter(value => isVisible(value.Source));
    };

    const toggleFilter = (source: string): void => {
        if (source === 'app') {
            setFilterApp(!filterApp);
        }
        if (source === 'action') {
            setFilterAction(!filterAction);
        }
        if (source === 'device') {
            setFilterDevice(!filterDevice);
        }
        if (source === 'status') {
            setFilterStatus(!filterStatus);
        }
        if (source === 'step') {
            setFilterStep(!filterStep);
        }
        if (source === 'testrunner') {
            setFilterTestrunner(!filterTestrunner);
        }
    };

    useEffect(() => {
        setFilteredEntries(filterEntries(entries));
    }, [filterApp, filterTestrunner, filterAction, filterStep, filterStatus, entries]);

    return (
        <div className={ classes.chip }>
            <Box sx={ { width: '100%', height: '50px' } }>
                <Chip className={ filterApp ? 'chip--app' : 'chip--app--unchecked' } label={ 'app' }
                    clickable={ true }
                    variant={ filterApp ? 'filled' : 'outlined' }
                    onClick={ () => toggleFilter('app') }/>
                <Chip className={ filterStep ? 'chip--step' : 'chip--step--unchecked' } label={ 'step' }
                    clickable={ true }
                    variant={ filterStep ? 'filled' : 'outlined' }
                    onClick={ () => toggleFilter('step') }/>
                <Chip className={ filterDevice ? 'chip--device' : 'chip--device--unchecked' }
                    label={ 'device' }
                    clickable={ true }
                    variant={ filterDevice ? 'filled' : 'outlined' }
                    onClick={ () => toggleFilter('device') }/>
                <Chip className={ filterStatus ? 'chip--status' : 'chip--status--unchecked' }
                    label={ 'status' }
                    clickable={ true }
                    variant={ filterStatus ? 'filled' : 'outlined' }
                    onClick={ () => toggleFilter('status') }/>
                <Chip className={ filterTestrunner ? 'chip--testrunner' : 'chip--testrunner--unchecked' }
                    label={ 'testrunner' } clickable={ true }
                    variant={ filterTestrunner ? 'filled' : 'outlined' }
                    onClick={ () => toggleFilter('testrunner') }/>
                <Chip className={ filterAction ? 'chip--action' : 'chip--action--unchecked' }
                    label={ 'action' } clickable={ true }
                    variant={ filterAction ? 'filled' : 'outlined' }
                    onClick={ () => toggleFilter('action') }/>
            </Box>
            <DataGrid
                autoHeight={ true }
                getRowId={ (row) => row.ID }
                rows={ filteredEntries }
                columns={ columns }
                checkboxSelection={ false }
                disableSelectionOnClick={ true }
                disableColumnFilter={ true }
            />
        </div>
    );
};

export default ProtocolLogComponent;
