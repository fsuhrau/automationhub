import React, {useState} from 'react';
import Grid from '@mui/material/Grid';
import {IAppData} from '../../types/app';
import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    MenuItem,
    Select,
    Slide,
    TextField,
    Typography
} from '@mui/material';
import {getPlatformTypes, PlatformType} from "../../types/platform.type.enum";
import {TransitionProps} from "@mui/material/transitions";

const Transition = React.forwardRef(function Transition(
    props: TransitionProps & {
        children: React.ReactElement;
    },
    ref: React.Ref<unknown>,
) {
    return <Slide direction="left" ref={ref} {...props} />;
});

export interface ApplicationProps {
    open: boolean;
    onSubmit: (data: IAppData) => void;
    onClose: () => void;
}

const NewAppDialog: React.FC<ApplicationProps> = (props: ApplicationProps) => {

    const {open, onClose, onSubmit} = props;

    const platformTypes = getPlatformTypes();

    const [state, setState] = useState<IAppData>({
        name: '',
        identifier: '',
        platform: PlatformType.Android,
    } as IAppData)

    const handleClose = () => {
        onClose();
    };

    const submit = () => {
        onSubmit(state);
        handleClose();
    }

    return (
        <Dialog
            open={open}
            onClose={handleClose}
            fullWidth={true}
        >
            <DialogTitle>
                <Typography variant={"h5"}>Add a new App</Typography>
            </DialogTitle>
            <DialogContent>
                <Grid container={true} sx={{padding: 2}}>
                    <Grid size={12}>
                        <Typography variant={"caption"}>App Name</Typography>
                    </Grid>
                    <Grid size={12}>
                        <TextField
                            hiddenLabel
                            id="appname"
                            placeholder="Name"
                            value={state.name}
                            size="small"
                            fullWidth={true}
                            onChange={event => setState(prevState => ({...state, name: event.target.value}))}
                        />
                    </Grid>
                    <Grid size={12}>
                        <Typography variant={"caption"}>Platform</Typography>
                    </Grid>
                    <Grid size={12}>
                        <Select
                            fullWidth={true}
                            defaultValue={state.platform}
                            labelId="platform-type-selection"
                            id="platform-type"
                            label="Platform"
                            onChange={event => setState(prevState => ({
                                ...prevState,
                                platform: +event.target.value as PlatformType
                            }))}
                        >
                            {platformTypes.map((value) => (
                                <MenuItem key={'tt_' + value.id}
                                          value={value.id}>{value.name}</MenuItem>
                            ))}
                        </Select>
                    </Grid>
                    <Grid size={12}>
                        <Typography variant={"caption"}>Bundle Identifier</Typography>
                    </Grid>
                    <Grid size={12}>
                        <TextField
                            fullWidth={true}
                            hiddenLabel
                            id="appidentifier"
                            placeholder="com.example.app"
                            value={state.identifier}
                            size="small"
                            onChange={event => setState(prevState => ({...state, identifier: event.target.value}))}
                        />
                    </Grid>
                </Grid>
            </DialogContent>
            <DialogActions>
                <Button variant={'outlined'} onClick={onClose}>Close</Button>
                <Button variant={"contained"} onClick={submit}>Create</Button>
            </DialogActions>
        </Dialog>
    );
};

export default NewAppDialog;
