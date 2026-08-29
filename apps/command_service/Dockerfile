FROM mcr.microsoft.com/dotnet/sdk:8.0 AS builder

WORKDIR /src

COPY schemas/api/ schemas/api/
COPY apps/command_service/ apps/command_service/

WORKDIR /src/apps/command_service

RUN dotnet publish -c Release -o /out

FROM mcr.microsoft.com/dotnet/aspnet:8.0

WORKDIR /app

COPY --from=builder /out .

EXPOSE 50051

ENTRYPOINT ["dotnet", "CommandService.dll"]
