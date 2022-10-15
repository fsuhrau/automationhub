import React, { useEffect, useState } from 'react';
import { Box, Button, Card, CardMedia, Chip, Popover, Popper, TextField, Typography } from '@mui/material';
import IProtocolEntryData from '../types/protocol.entry';
import { DataGrid, GridCellValue, GridColDef, GridRenderCellParams } from '@mui/x-data-grid';
import { makeStyles } from '@mui/styles';
import CellExpand from './cell.expand.component';
import Grid from "@mui/material/Grid";

const useStyles = makeStyles(theme => ({
    chip: {
        '& .chip--error': {
            backgroundColor: '#DB3A34',
            color: '#ffffff',
            margin: '5px',
        },
        '& .chip--error--unchecked': {
            backgroundColor: '#ffffff',
            fontcolor: '#DB3A34',
            margin: '5px',
        },
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

    const timeFrom = (value: GridCellValue): string => {
        return new Date((value as number) * 1000).toISOString().substr(11, 8);
    };

    const nanosFrom = (value: GridCellValue): string => {
        const str = (value as number).toFixed(4);
        return str.substring(str.length - 4);
    };

    const renderCellExpand = (params: GridRenderCellParams): React.ReactNode => {
        return (
            <CellExpand id={params.row.ID} value={params.value} data={params.row.Data} />
        );
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
            renderCell: renderCellExpand,
        },
    ];

    type FilterType = {
        Errors: boolean,
        App: boolean,
        Action: boolean,
        Device: boolean,
        Status: boolean,
        Step: boolean,
        TestRunner: boolean,
        Content: string,
    };

    const [filter, setFilter] = useState<FilterType>({
        Errors: false,
        App: true,
        Action: true,
        Device: true,
        Status: true,
        Step: true,
        TestRunner: true,
        Content: "",
    })

    const isVisible = (source: string): boolean => {
        return (filter.App && source === 'app') ||
            (filter.Action && source === 'action') ||
            (filter.Device && source === 'device') ||
            (filter.Status && source === 'status') ||
            (filter.Step && source === 'step') ||
            (filter.TestRunner && source === 'testrunner' ||
                source === 'screen');
    };

    const filterEntries = entries.filter(value => (!filter.Errors && isVisible(value.Source) || (value.Level == 'error')) && (filter.Content.length < 2 || value.Message.indexOf(filter.Content) !== -1));

    return (
        <Grid container={true} className={ classes.chip } sx={{padding: 1}}>
            <Grid item={true} xs={12} container={true} justifyContent={"center"} sx={{padding: 1}}>
                <Chip className={ filter.Errors ? 'chip--error' : 'chip--error--unchecked' } label={ 'errors' }
                      clickable={ true }
                      variant={ filter.Errors ? 'filled' : 'outlined' }
                      onClick={ () => setFilter(prevState => ({...prevState, Errors: !prevState.Errors})) }/>
                <Chip className={ filter.App ? 'chip--app' : 'chip--app--unchecked' } label={ 'app' }
                      clickable={ true }
                      variant={ filter.App ? 'filled' : 'outlined' }
                      onClick={ () => setFilter(prevState => ({...prevState, App: !prevState.App})) }/>
                <Chip className={ filter.Step ? 'chip--step' : 'chip--step--unchecked' } label={ 'step' }
                      clickable={ true }
                      variant={ filter.Step ? 'filled' : 'outlined' }
                      onClick={ () => setFilter(prevState => ({...prevState, Step: !prevState.Step})) }/>
                <Chip className={ filter.Device ? 'chip--device' : 'chip--device--unchecked' }
                      label={ 'device' }
                      clickable={ true }
                      variant={ filter.Device ? 'filled' : 'outlined' }
                      onClick={ () => setFilter(prevState => ({...prevState, Device: !prevState.Device})) }/>
                <Chip className={ filter.Status ? 'chip--status' : 'chip--status--unchecked' }
                      label={ 'status' }
                      clickable={ true }
                      variant={ filter.Status ? 'filled' : 'outlined' }
                      onClick={ () => setFilter(prevState => ({...prevState, Status: !prevState.Status})) }/>
                <Chip className={ filter.TestRunner ? 'chip--testrunner' : 'chip--testrunner--unchecked' }
                      label={ 'testrunner' } clickable={ true }
                      variant={ filter.TestRunner ? 'filled' : 'outlined' }
                      onClick={ () => setFilter(prevState => ({...prevState, TestRunner: !prevState.TestRunner})) }/>
                <Chip className={ filter.Action ? 'chip--action' : 'chip--action--unchecked' }
                      label={ 'action' } clickable={ true }
                      variant={ filter.Action ? 'filled' : 'outlined' }
                      onClick={ () => setFilter(prevState => ({...prevState, Action: !prevState.Action})) }/>
                <TextField
                    id="textfilter"
                    label="Content"
                    size={"small"}
                    value={filter.Content}
                    onChange={(e)=> setFilter(prevState => ({...prevState, Content: e.target.value}))} />
            </Grid>
            <Grid item={true} xs={12}>
                <DataGrid
                    autoHeight={ true }
                    getRowId={ (row) => row.ID }
                    rows={ filterEntries }
                    columns={ columns }
                    checkboxSelection={ false }
                    disableSelectionOnClick={ true }
                    disableColumnFilter={ true }
                />

            </Grid>
        </Grid>
    );
};

export default ProtocolLogComponent;
