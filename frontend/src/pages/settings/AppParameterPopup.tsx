import React, {useState} from "react";
import {Dialog, DialogActions, DialogContent, MenuItem, TextField, Typography} from "@mui/material";
import Button from "@mui/material/Button";
import {AppParameter, AppParameterOption, AppParameterString, AppParameterType, Parameter} from "../../types/app";
import Select from "@mui/material/Select";
import Grid from "@mui/material/Grid";

interface AppParameterPopupProps {
    open: boolean,
    parameter: AppParameter | null,
    onClose: () => void,
    onSubmit: (param: AppParameter) => void,
}

const AppParameterPopup: React.FC<AppParameterPopupProps> = (props: AppParameterPopupProps) => {

    type State = {
        name: string,
        type: AppParameterType
        defaultValue: string
        options: string[],
    }

    const initialState = (param: AppParameter | null) : State => {
        if (param == null) {
            return {
                name: '',
                type: 'string',
                defaultValue: '',
                options: [],
            }
        }
        let defaultValue = '';
        let type = param.type.type;
        let options: string[] = [];

        if (param.type.type === 'string') {
            const stringParam = (param.type as AppParameterString);
            defaultValue = stringParam.defaultValue;
        }
        if (param.type.type === 'option') {
            const optionParam = (param.type as AppParameterOption);
            defaultValue = optionParam.defaultValue;
            options = optionParam.options;
        }

        return {
            name: param.name,
            type: type,
            defaultValue: defaultValue,
            options: options,
        }
    }

    const [state, setState] = useState<State>(initialState(props.parameter));

    const handleEditClose = () => {
        props.onClose();
    };

    const handleEditSubmit = () => {
        if (props.parameter) {
            props.parameter.name = state.name;

            var param: Parameter = {} as Parameter;
            if (props.parameter.type.type == 'string') {
                param = props.parameter.type as AppParameterString;
            }
            if (props.parameter.type.type == 'option') {
                param = props.parameter.type as AppParameterOption;
                (param as AppParameterOption).options = state.options
            }
            param!.defaultValue = state.defaultValue;
            props.parameter.type = param;
            props.onSubmit(props.parameter)
            return;
        }
        var type = {}
        switch (state.type) {
            case 'string':
                type = {type: state.type, defaultValue: state.defaultValue}
                break
            case 'option':
                type = {type: state.type, defaultValue: state.defaultValue, options: state.options}
                break
        }
        props.onSubmit({name: state.name, type: type as Parameter} as AppParameter)
        handleEditClose();
    };

    const handleOptionChange = (idx: number, e: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) => {
        setState(prevState => ({
            ...prevState,
            options: prevState.options.map((d, i) => i === idx ? e.target.value : d)
        }))
    };

    const handleOptionRemove = (idx: number) => {
        setState(prevState => ({
            ...prevState,
            options: prevState.options.filter((d, i) => i != idx)
        }))
    };

    const handleOptionAdd = () => {
        setState(prevState => ({
            ...prevState,
            options: [...prevState.options, ''],
        }))
    };

    React.useEffect(() => {
        setState(initialState(props.parameter))
    }, [props.parameter]);

    return (
        <Dialog open={props.open} onClose={handleEditClose}>
            <DialogContent>
                <Grid container={true} spacing={1}>
                    <Grid size={12}>
                        <Typography variant={"body2"}>Name</Typography>
                        <TextField placeholder={"Name"}
                                   value={state.name}
                                   fullWidth={true}
                                   onChange={e => setState(prevState => ({
                                       ...prevState,
                                       name: e.target.value
                                   }))}/>
                    </Grid>
                    <Grid size={12}>
                        <Typography variant={"body2"}>Type</Typography>
                        <Select
                            fullWidth={true}
                            labelId="env-input-type-select-label"
                            id="env-input-type-select"
                            value={state.type}
                            label="Input Type"
                            onChange={e => setState(prevState => ({
                                ...prevState,
                                type: e.target.value as AppParameterType
                            }))}
                        >
                            <MenuItem value={'string'}>String</MenuItem>
                            <MenuItem value={'option'}>Option</MenuItem>
                        </Select>
                    </Grid>
                    {state.type == 'option' && <Grid container={true} size={12}>
                        <Typography variant={"body2"}>Options</Typography>
                        {(state.options.map((o, idx) => <Grid
                            key={`param_option_${idx}`}
                            container={true} size={12}>
                            <Grid size={10}>
                                <TextField fullWidth={true} value={o} onChange={e => handleOptionChange(idx, e)}/>
                            </Grid>
                            <Grid size={2}>
                                <Button color={"error"} variant={"contained"}
                                        onClick={() => handleOptionRemove(idx)}>Remove</Button>
                            </Grid>
                        </Grid>))}
                        <Grid size={10}>
                        </Grid>
                        <Grid size={2}>
                            <Button variant={"contained"} onClick={handleOptionAdd}>Add</Button>
                        </Grid>
                    </Grid>}
                    <Grid size={12}>
                        <Typography variant={"body2"}>Default Value</Typography>
                        <TextField fullWidth={true} placeholder={"DefaultValue"} onChange={e => setState(prevState => ({
                            ...prevState,
                            defaultValue: e.target.value
                        }))}/>
                    </Grid>
                </Grid>
            </DialogContent>
            <DialogActions>
                <Button onClick={handleEditClose}>Cancel</Button>
                <Button variant={'contained'} onClick={handleEditSubmit}>Save</Button>
            </DialogActions>
        </Dialog>
    )
}

export default AppParameterPopup;