using Grpc.Core;
using RemoteControlCommand.V1;

public class RemoteControlCommandGrpcService : RemoteControlCommandService.RemoteControlCommandServiceBase
{
    private readonly CommandStore _store;

    public RemoteControlCommandGrpcService(CommandStore store)
    {
        _store = store;
    }

    public override Task<CommandResult> Send(SendCommandRequest request, ServerCallContext context)
    {
        return Task.FromResult(_store.Send(request));
    }

    public override Task<CommandResult> GetResult(CommandId request, ServerCallContext context)
    {
        return Task.FromResult(_store.GetResult(request.Id));
    }

    public override Task<CommandResult> Cancel(CommandId request, ServerCallContext context)
    {
        throw new RpcException(new Status(StatusCode.Unimplemented, "not implemented in mvp"));
    }

    public override Task<CommandResults> GetResultsBySensorId(SensorId request, ServerCallContext context)
    {
        throw new RpcException(new Status(StatusCode.Unimplemented, "not implemented in mvp"));
    }
}
