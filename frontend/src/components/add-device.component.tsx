import {ChangeEvent, Component} from "react";
import IDeviceData from "../types/device";
import DeviceDataService from "../services/device.service";

type Props = {};

type State = IDeviceData & {
    submitted: boolean
};

export default class AddDevice extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.onChangeName = this.onChangeName.bind(this);
        this.saveDevice = this.saveDevice.bind(this);
        this.newDevice = this.newDevice.bind(this);

        this.state = {
            OS: "",
            id: null,
            DeviceIdentifier: "",
            DeviceType: 0,
            Name: "",
            RAM: 0,
            SoC: "",
            DisplaySize: "",
            DPI: 0,
            OSVersion: "",
            GPU: "",
            ABI: "",
            OpenGL_ES_Version: 0,
            Status: 0,
            Manager: "",
            submitted: false
        };
    }

    onChangeName(e: ChangeEvent<HTMLInputElement>) {
        this.setState({
            Name: e.target.value
        })
    }

    saveDevice() {
        const data: IDeviceData = {
            OS: "",
            Name: this.state.Name,
            DeviceIdentifier: this.state.DeviceIdentifier,
            ABI: "",
            DPI: 0,
            DeviceType: 0,
            DisplaySize: "",
            GPU: "",
            Manager: "",
            OSVersion: "",
            OpenGL_ES_Version: 0,
            RAM: 0,
            SoC: "",
            Status: 0
        }

        DeviceDataService.create(data)
            .then(response => {
                this.setState({
                    id: response.data.id,
                });
                console.log(response.data);
            }).catch(e => {
            console.log(e);
        });
    }

    newDevice() {
        this.setState({
            id: null,
            Name: this.state.Name,
            DeviceIdentifier: this.state.DeviceIdentifier,
            ABI: "",
            DPI: 0,
            DeviceType: 0,
            DisplaySize: "",
            GPU: "",
            Manager: "",
            OS: "",
            OSVersion: "",
            OpenGL_ES_Version: 0,
            RAM: 0,
            SoC: "",
            Status: 0,
        })
    }

    render() {
        const {submitted, Name} = this.state;
        return (
            <div className="submit-form">
                {submitted ? (
                    <div>
                        <h4>You submitted successfully!</h4>
                        <button className="btn btn-success" onClick={this.newDevice}>
                            Add
                        </button>
                    </div>
                ) : (
                    <div>
                        <div className="form-group">
                            <label htmlFor="title">Title</label>
                            <input
                                type="text"
                                className="form-control"
                                id="title"
                                required
                                value={Name}
                                onChange={this.onChangeName}
                                name="title"
                            />
                        </div>

                        <button onClick={this.saveDevice} className="btn btn-success">
                            Submit
                        </button>
                    </div>
                )}
            </div>
        );
    }
}
