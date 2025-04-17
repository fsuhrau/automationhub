import React, {useRef, useState} from 'react';
import {
    Checkbox,
    Chip,
    Divider,
    FormControlLabel,
    FormGroup,
    IconButton,
    InputBase,
    Popover
} from '@mui/material';
import { makeStyles } from '@mui/styles';
import IProtocolEntryData from '../types/protocol.entry';
import {
    DataGrid,
    gridClasses,
    GridColDef,
    GridRenderCellParams,
    GridToolbarColumnsButton,
    GridToolbarContainer,
    GridToolbarExport
} from '@mui/x-data-grid';
import CellExpand from './cell.expand.component';
import Grid from "@mui/material/Grid";
import SearchIcon from '@mui/icons-material/Search';
import {ClearIcon} from "@mui/x-date-pickers";
import Paper from "@mui/material/Paper";
import MenuIcon from '@mui/icons-material/Menu';

const useStyles = makeStyles({
    dataGrid: {
        '& .MuiDataGrid-cell': {
            whiteSpace: 'normal !important',
            wordWrap: 'break-word !important',
            lineHeight: '1.5 !important',
            display: 'block !important',
        },
    },
    chip: {
        '& .chip--error': {
            backgroundColor: '#DB3A34',
            margin: '5px',
            '& .MuiChip-label': {
                color: 'white', // Set the desired color
            },
        },
        '& .chip--app': {
            backgroundColor: '#177E89',
            margin: '5px',
            '& .MuiChip-label': {
                color: 'white', // Set the desired color
            },
        },
        '& .chip--testrunner': {
            backgroundColor: '#084C61',
            margin: '5px',
            '& .MuiChip-label': {
                color: 'white', // Set the desired color
            },
        },
        '& .chip--step': {
            backgroundColor: '#DB3A34',
            margin: '5px',
            '& .MuiChip-label': {
                color: 'white', // Set the desired color
            },
        },
        '& .chip--status': {
            backgroundColor: '#FFC857',
            margin: '5px',
            '& .MuiChip-label': {
                color: 'black', // Set the desired color
            },
        },
        '& .chip--device': {
            backgroundColor: '#323031',
            margin: '5px',
            '& .MuiChip-label': {
                color: 'white', // Set the desired color
            },
        },
        '& .chip--action': {
            backgroundColor: '#323031',
            margin: '5px',
            '& .MuiChip-label': {
                color: 'white', // Set the desired color
            },
        },
    },
});

interface TestProtocolContentProps {
    entries: IProtocolEntryData[]
}

const ProtocolLogComponent: React.FC<TestProtocolContentProps> = (props: TestProtocolContentProps) => {
    const classes = useStyles();
    const {entries} = props;

    const timeFrom = (value: number): string => {
        return new Date(value * 1000).toISOString().substr(11, 8);
    };

    const nanosFrom = (value: number): string => {
        const str = value.toFixed(4);
        return str.substring(str.length - 4);
    };

    const renderCellExpand = (params: GridRenderCellParams): React.ReactNode => {
        return (
            <CellExpand id={params.row.ID} value={params.value} data={params.row.Data}/>
        );
    };

    const columns: GridColDef[] = [
        {
            field: 'Runtime',
            headerName: 'Time',
            width: 120,
            sortable: true,
            filterable: false,
            disableColumnMenu: true,
            renderCell: (params) => {
                return (<div>{timeFrom(params.value)}.{nanosFrom(params.value)}</div>);
            },
        },
        {
            field: 'Source',
            headerName: 'Source',
            width: 110,
            sortable: false,
            filterable: false,
            disableColumnMenu: true,
            renderCell: (params) => {
                return (<Chip className={`chip--${params.value}`} label={params.value}/>);
            },
        },
        {
            field: 'Level',
            headerName: 'Level',
            width: 60,
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

    type SourceType = {
        Errors: boolean,
        App: boolean,
        Action: boolean,
        Device: boolean,
        Status: boolean,
        Step: boolean,
        TestRunner: boolean,
        Content: string,
    };

    const [source, setSource] = useState<SourceType>({
        Errors: false,
        App: true,
        Action: true,
        Device: true,
        Status: true,
        Step: true,
        TestRunner: true,
        Content: "",
    });

    const isVisible = (value: string): boolean => {
        return (source.App && value === 'app') ||
            (source.Action && value === 'action') ||
            (source.Device && value === 'device') ||
            (source.Status && value === 'status') ||
            (source.Step && value === 'step') ||
            (source.TestRunner && value === 'testrunner' ||
                value === 'screen');
    };

    const filterEntries = entries.filter(value => (!source.Errors && isVisible(value.Source) || (value.Level == 'error')) && (source.Content.length < 2 || value.Message.indexOf(source.Content) !== -1));

    const anchorRef = useRef<HTMLButtonElement | null>(null);
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl);
    const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
        setAnchorEl(event.currentTarget);
    };
    const handleClose = () => {
        setAnchorEl(null);
    };

    const handleCheckboxChange = (filterKey: keyof SourceType) => {
        setSource(prevState => ({...prevState, [filterKey]: !prevState[filterKey]}));
    };

    function CustomToolbar() {
        return (
            <GridToolbarContainer>
                <GridToolbarColumnsButton/>
                <Paper
                    component="form"
                    sx={{ p: '2px 4px', display: 'flex', alignItems: 'center', flexGrow: 1 }}
                >
                    <IconButton sx={{ p: '10px' }} aria-label="menu"
                                id="row-filter-button"
                                aria-controls={open ? 'row-filter-popover' : undefined}
                                aria-haspopup="true"
                                aria-expanded={open ? 'true' : undefined}
                                onClick={handleClick}
                                ref={anchorRef}
                    >
                        <MenuIcon />
                    </IconButton>
                    <Popover
                        id="row-filter-popover"
                        anchorEl={anchorRef.current}
                        open={open}
                        onClose={handleClose}
                        anchorOrigin={{
                            vertical: 'bottom',
                            horizontal: 'left',
                        }}
                    >
                        <FormGroup sx={{padding: 1}}>
                            <FormControlLabel
                                control={<Checkbox checked={source.Errors}
                                                   onChange={() => handleCheckboxChange('Errors')}/>}
                                label="Only Errors"
                            />
                            <Divider />
                            <FormControlLabel
                                control={<Checkbox checked={source.App} onChange={() => handleCheckboxChange('App')}/>}
                                label="App"
                            />
                            <FormControlLabel
                                control={<Checkbox checked={source.Step} onChange={() => handleCheckboxChange('Step')}/>}
                                label="Step"
                            />
                            <FormControlLabel
                                control={<Checkbox checked={source.Device}
                                                   onChange={() => handleCheckboxChange('Device')}/>}
                                label="Device"
                            />
                            <FormControlLabel
                                control={<Checkbox checked={source.Status}
                                                   onChange={() => handleCheckboxChange('Status')}/>}
                                label="Status"
                            />
                            <FormControlLabel
                                control={<Checkbox checked={source.TestRunner}
                                                   onChange={() => handleCheckboxChange('TestRunner')}/>}
                                label="TestRunner"
                            />
                            <FormControlLabel
                                control={<Checkbox checked={source.Action}
                                                   onChange={() => handleCheckboxChange('Action')}/>}
                                label="Action"
                            />
                        </FormGroup>
                    </Popover>
                    <InputBase
                        sx={{ ml: 1, flex: 1 }}
                        placeholder="Filter log"
                        inputProps={{ 'aria-label': 'Filter log' }}
                        value={source.Content}
                        autoFocus={true}
                        onChange={(e) => setSource(prevState => ({...prevState, Content: e.currentTarget.value}))}
                    />
                    <IconButton type="button" sx={{ p: '10px' }} aria-label="search"
                    >
                        <SearchIcon />
                    </IconButton>
                    <Divider sx={{ height: 28, m: 0.5 }} orientation="vertical" />
                    <IconButton color="primary" sx={{ p: '10px' }} aria-label="clear"
                                onClick={(e) => setSource(prevState => ({...prevState, Content: ''}))}>
                        <ClearIcon/>
                    </IconButton>
                </Paper>
                <GridToolbarExport
                    slotProps={{
                        tooltip: {title: 'Export data'},
                        button: {variant: 'outlined'},
                    }}
                />
            </GridToolbarContainer>
        );
    }

    return (
        <Grid container={true} size={12} className={classes.chip}>
            <DataGrid
                density={"compact"}
                getRowId={(row) => row.ID}
                rows={filterEntries}
                columns={columns}
                checkboxSelection={false}
                disableRowSelectionOnClick={true}
                disableColumnFilter={true}
                getRowHeight={() => 'auto'}
                className={classes.dataGrid}
                sx={{
                    [`& .${gridClasses.cell}`]: {
                        py: 1,
                    },
                }}
                slots={{
                    toolbar: CustomToolbar
                }}
            />
        </Grid>
    );
};

export default ProtocolLogComponent;