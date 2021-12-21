import React, { FC, ReactElement, useEffect } from 'react';
import Grid from '@mui/material/Grid';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Checkbox from '@mui/material/Checkbox';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import IDeviceData from '../types/device';
import { getAllDevices } from '../services/device.service';
import { Typography } from '@mui/material';
import { makeStyles } from '@mui/styles';

const useStyles = makeStyles(theme => ({
    root: {
        margin: 'auto',
    },
    paper: {
        width: 200,
        height: 230,
        overflow: 'auto',
    },
    button: {
        margin: '0px',
    },
}));

interface DeviceSelectionProps {
    onSelectionChanged: (devices: IDeviceData[]) => void;
    selectedDevices: IDeviceData[];
}

function not(a: IDeviceData[], b: IDeviceData[]): IDeviceData[] {
    return a.filter((value) => b.indexOf(value) === -1);
}

function intersection(a: IDeviceData[], b: IDeviceData[]): IDeviceData[] {
    return a.filter((value) => b.indexOf(value) !== -1);
}

const DeviceSelection: FC<DeviceSelectionProps> = (props) => {
    const classes = useStyles();
    const { selectedDevices, onSelectionChanged } = props;

    const [checked, setChecked] = React.useState<IDeviceData[]>([]);

    const [left, setLeft] = React.useState<IDeviceData[]>([]);
    const [right, setRight] = React.useState<IDeviceData[]>([]);

    const leftChecked = intersection(checked, left);
    const rightChecked = intersection(checked, right);

    const handleToggle = (value: IDeviceData) => () => {
        const currentIndex = checked.indexOf(value);
        const newChecked = [...checked];

        if (currentIndex === -1) {
            newChecked.push(value);
        } else {
            newChecked.splice(currentIndex, 1);
        }

        setChecked(newChecked);
    };

    const handleAllRight = (): void => {
        setRight(right.concat(left));
        setLeft([]);
    };

    const handleCheckedRight = (): void => {
        setRight(right.concat(leftChecked));
        setLeft(not(left, leftChecked));
        setChecked(not(checked, leftChecked));
    };

    const handleCheckedLeft = (): void => {
        setLeft(left.concat(rightChecked));
        setRight(not(right, rightChecked));
        setChecked(not(checked, rightChecked));
    };

    const handleAllLeft = (): void => {
        setLeft(left.concat(right));
        setRight([]);
    };

    useEffect(() => {
        getAllDevices().then(response => {
            const r = response.data.filter(value => selectedDevices.find(element => element.ID == value.ID));
            setRight(r);
            setLeft(not(response.data, r));
        }).catch(e => {
        });
    }, []);

    useEffect(() => {
        onSelectionChanged(right);
    }, [right, onSelectionChanged]);

    const customList = (items: IDeviceData[]): ReactElement => (
        <Paper className={classes.paper}>
            <List dense={true} component="div" role="list">
                {items.map((value) => {
                    const labelId = `transfer-list-item-${value.ID}-label`;

                    return (
                        <ListItem key={value.ID} role="listitem" button={true} onClick={handleToggle(value)}>
                            <ListItemIcon>
                                <Checkbox
                                    checked={checked.indexOf(value) !== -1}
                                    tabIndex={-1}
                                    disableRipple={true}
                                    inputProps={{ 'aria-labelledby': labelId }}
                                />
                            </ListItemIcon>
                            <ListItemText id={labelId} primary={`${value.DeviceIdentifier}(${value.Name})`} />
                        </ListItem>
                    );
                })}
                <ListItem />
            </List>
        </Paper>
    );

    return (
        <Grid
            container={true}
            spacing={2}
            justifyContent="center"
            alignItems="center"
            className={classes.root}
        >
            <Grid item={true}>
                <Typography variant={'subtitle1'}>Available Devices</Typography>
                {customList(left)}
            </Grid>
            <Grid item={true}>
                <Grid container={true} direction="column" alignItems="center">
                    <Button
                        variant="outlined"
                        size="small"
                        className={classes.button}
                        onClick={handleAllRight}
                        disabled={left.length === 0}
                        aria-label="move all right"
                    >
                        ≫
                    </Button>
                    <Button
                        variant="outlined"
                        size="small"
                        className={classes.button}
                        onClick={handleCheckedRight}
                        disabled={leftChecked.length === 0}
                        aria-label="move selected right"
                    >
                        &gt;
                    </Button>
                    <Button
                        variant="outlined"
                        size="small"
                        className={classes.button}
                        onClick={handleCheckedLeft}
                        disabled={rightChecked.length === 0}
                        aria-label="move selected left"
                    >
                        &lt;
                    </Button>
                    <Button
                        variant="outlined"
                        size="small"
                        className={classes.button}
                        onClick={handleAllLeft}
                        disabled={right.length === 0}
                        aria-label="move all left"
                    >
                        ≪
                    </Button>
                </Grid>
            </Grid>
            <Grid item={true}>
                <Typography variant={'subtitle1'}>Selected Devices</Typography>
                {customList(right)}
            </Grid>
        </Grid>
    );
};

export default DeviceSelection;
