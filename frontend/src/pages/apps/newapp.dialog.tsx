import React, { useState } from 'react';
import Grid from '@mui/material/Grid';
import { IAppData } from '../../types/app';
import {
    Button,
    Dialog,
    IconButton,
    MenuItem,
    Select,
    Slide,
    TextField,
    Typography
} from '@mui/material';
import { getPlatformTypes, PlatformType } from "../../types/platform.type.enum";
import { TransitionProps } from "@mui/material/transitions";
import { Close } from "@mui/icons-material";

const Transition = React.forwardRef(function Transition(
    props: TransitionProps & {
        children: React.ReactElement;
    },
    ref: React.Ref<unknown>,
) {
    return <Slide direction="left" ref={ ref } { ...props } />;
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
        Name: '',
        Identifier: '',
        Platform: PlatformType.Android,
        DefaultParameter: ''
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
            fullScreen
            open={ open }
            onClose={ handleClose }
            TransitionComponent={ Transition }
        >
            <Grid container={ true } sx={ {padding: 5} }>
                <Grid size={12}>
                    <Typography variant={ "h5" }><IconButton onClick={ onClose }><Close/></IconButton>Create a new
                        App</Typography>
                </Grid>
                <Grid size={2} container={ true } sx={ {padding: 3} } spacing={ 2 }>
                    <Grid size={ 12 }>
                        <Typography variant={ "caption" }>App Name</Typography>
                    </Grid>
                    <Grid size={ 12 }>
                        <TextField
                            hiddenLabel
                            id="appname"
                            placeholder="Name"
                            variant="filled"
                            value={ state.Name }
                            size="small"
                            fullWidth={true}
                            onChange={ event => setState(prevState => ({...state, Name: event.target.value})) }
                        />
                    </Grid>
                    <Grid size={ 12 }>
                        <Typography variant={ "caption" }>Platform</Typography>
                    </Grid>
                    <Grid size={ 12 }>
                        <Select
                            fullWidth={true}
                            defaultValue={ state.Platform }
                            labelId="platform-type-selection"
                            id="platform-type"
                            label="Platform"
                            onChange={ event => setState(prevState => ({
                                ...prevState,
                                Platform: +event.target.value as PlatformType
                            })) }
                        >
                            { platformTypes.map((value) => (
                                <MenuItem key={ 'tt_' + value.id }
                                          value={ value.id }>{ value.name }</MenuItem>
                            )) }
                        </Select>
                    </Grid>
                    <Grid size={ 12 }>
                        <Typography variant={ "caption" }>Bundle Identifier</Typography>
                    </Grid>
                    <Grid size={ 12 }>
                        <TextField
                            fullWidth={true}
                            hiddenLabel
                            id="appidentifier"
                            placeholder="com.example.app"
                            variant="filled"
                            value={ state.Identifier }
                            size="small"
                            onChange={ event => setState(prevState => ({...state, Identifier: event.target.value})) }
                        />
                    </Grid>
                    <Grid size={ 12 }>
                        <Typography variant={ "caption" }>Default Parameter</Typography>
                    </Grid>
                    <Grid size={ 12 }>
                        <TextField
                            fullWidth={true}
                            hiddenLabel
                            id="appdefaultparameter"
                            variant="filled"
                            placeholder={"Proxy=http://localhost:8888"}
                            value={ state.DefaultParameter }
                            size="small"
                            onChange={ event => setState(prevState => ({
                                ...state,
                                DefaultParameter: event.target.value
                            })) }
                        />
                    </Grid>
                    <Grid size={ 12 }>
                        <Button variant={"outlined"} onClick={submit}>Create</Button>
                    </Grid>
                </Grid>
            </Grid>
        </Dialog>
    );
};

export default NewAppDialog;
