export interface INodeStatus {

}

export interface INodeData {
    id: number,
    identifier: string,
    name: string,
    status: number,
    //status?: INodeStatus,
}

export const parseNodeData = (json: any): INodeData => {
    return {
        id: json.id,
        identifier: json.identifier,
        name: json.name,
        status: json.status,
    }
}
