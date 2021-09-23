import { ChangeEvent, Component, ReactElement } from 'react';
import IDeviceData from '../types/device';
import { createDevice } from '../services/device.service';

type Props = Record<string, never>;

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
            CompanyID: 0,
            Dev: undefined,
            OS: '',
            ID: null,
            DeviceIdentifier: '',
            DeviceType: 0,
            Name: '',
            RAM: 0,
            SOC: '',
            DisplaySize: '',
            DPI: 0,
            OSVersion: '',
            GPU: '',
            ABI: '',
            OpenGLESVersion: 0,
            Status: 0,
            Manager: '',
            submitted: false,
        };
    }

    onChangeName(e: ChangeEvent<HTMLInputElement>): void {
        this.setState({
            Name: e.target.value,
        });
    }

    saveDevice(): void {
        const data: IDeviceData = {
            CompanyID: 0,
            OpenGLESVersion: 0,
            OS: '',
            Name: this.state.Name,
            DeviceIdentifier: this.state.DeviceIdentifier,
            ABI: '',
            DPI: 0,
            DeviceType: 0,
            DisplaySize: '',
            GPU: '',
            Manager: '',
            OSVersion: '',
            RAM: 0,
            SOC: '',
            Status: 0,
        };

        createDevice(data)
            .then(response => {
                this.setState({
                    ID: response.data.ID,
                });
                console.log(response.data);
            }).catch(e => {
                console.log(e);
            });
    }

    newDevice(): void {
        this.setState((prevState) => ({
            ID: null,
            Name: prevState.Name,
            DeviceIdentifier: prevState.DeviceIdentifier,
            ABI: '',
            DPI: 0,
            DeviceType: 0,
            DisplaySize: '',
            GPU: '',
            Manager: '',
            OS: '',
            OSVersion: '',
            RAM: 0,
            Status: 0,
        }));
    }

    render(): ReactElement {
        const { submitted, Name } = this.state;
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
                                required={true}
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
