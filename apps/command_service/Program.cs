using Microsoft.AspNetCore.Server.Kestrel.Core;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddGrpc();
builder.Services.AddSingleton<CommandStore>();
builder.WebHost.ConfigureKestrel(options =>
{
    options.ListenAnyIP(50051, listen => listen.Protocols = HttpProtocols.Http2);
});

var app = builder.Build();

app.MapGrpcService<RemoteControlCommandGrpcService>();

app.Run();
