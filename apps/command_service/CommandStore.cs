using System.Collections.Concurrent;
using Google.Protobuf.WellKnownTypes;
using Grpc.Core;
using RemoteControlCommand.V1;

public class CommandStore
{
    private readonly ConcurrentDictionary<string, CommandResult> _commandsById = new();
    private readonly ConcurrentDictionary<string, string> _idsByIdempotencyKey = new();

    public CommandResult Send(SendCommandRequest request)
    {
        if (request.IdempotencyKey.Length > 0 && _idsByIdempotencyKey.TryGetValue(request.IdempotencyKey, out var existingId))
        {
            return _commandsById[existingId];
        }

        var command = new CommandResult
        {
            Id = Guid.NewGuid().ToString(),
            Status = CommandStatus.Pending,
            CreatedAt = Timestamp.FromDateTimeOffset(DateTimeOffset.UtcNow),
            InputPayload = request.InputPayload ?? new Struct(),
        };

        _commandsById[command.Id] = command;
        if (request.IdempotencyKey.Length > 0)
        {
            _idsByIdempotencyKey[request.IdempotencyKey] = command.Id;
        }

        _ = CompleteLaterAsync(command.Id);
        return command;
    }

    public CommandResult GetResult(string id)
    {
        if (!_commandsById.TryGetValue(id, out var command))
        {
            throw new RpcException(new Status(StatusCode.NotFound, $"command {id} not found"));
        }
        return command;
    }

    private async Task CompleteLaterAsync(string id)
    {
        await Task.Delay(TimeSpan.FromSeconds(2));
        if (_commandsById.TryGetValue(id, out var command))
        {
            var finished = command.Clone();
            finished.Status = CommandStatus.Done;
            finished.FinishedAt = Timestamp.FromDateTimeOffset(DateTimeOffset.UtcNow);
            finished.ResultData = new Struct();
            _commandsById[id] = finished;
        }
    }
}
