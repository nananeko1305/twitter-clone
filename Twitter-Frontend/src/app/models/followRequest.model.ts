export class FollowRequest {
    id: string = "";
    requester: string = "";
    receiver: string = "";
    status: number = 0;

    FollowRequest(id: string, requester: string, receiver: string, status: number) {
        this.id = id
        this.requester = requester;
        this.receiver = receiver;
        this.status = status;
    }
}