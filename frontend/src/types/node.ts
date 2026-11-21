export interface INodeData {
    id: number,
    identifier: string,
    name: string,
    status: number,
}

// the node as the master sends it
export type NodeJson = INodeData;

export const parseNodeData = (json: NodeJson): INodeData => {
    return {
        id: json.id,
        identifier: json.identifier,
        name: json.name,
        status: json.status,
    }
}
