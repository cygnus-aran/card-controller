import {
    AttributeTypeEnum,
    DynamoActions,
    EventsEnum,
    InputLambdaProps,
    KushkiStack,
    PatternEnum,
    PluginsEnum,
    ResourceEnum
} from "@kushki/cdk";
import {Schedule} from "aws-cdk-lib/aws-events";
import * as cdk from 'aws-cdk-lib';
import {Duration} from 'aws-cdk-lib';
import {AttributeType, StreamViewType} from "aws-cdk-lib/aws-dynamodb";
import {IResourceService} from "@kushki/cdk/lib/lib/repository/IResourceService";
import {SQSQueueResource} from "@kushki/cdk/lib/lib/repository/ResourceProps";
import {AccountEnvEnum} from "@kushki/cdk/lib/common/infraestructure/AccountEnvEnum";


const STACK: KushkiStack = new KushkiStack();
// Constants
const CODE_PATH: string = "./my-artifacts";
const HANDLER_PATH: string = "bootstrap";

const LAMBDA_PROPS = (
    functionName: string,
    handlerName: string
): InputLambdaProps => ({
    functionName,
    code: `${CODE_PATH}/${handlerName}.zip`,
    handler: `${HANDLER_PATH}`,
});

// INTERFACES
interface IVirtualPrivateCloud {
    securityGroups?: string[];
    vpcId?: string;
    vpcSubnets?: {
        subnets: string[];
    };
}

// Resources
const DYNAMO_BLOCKED_CARD = STACK.setResource({
    props: {
        partitionKey: {name: "cardID", type: AttributeType.STRING},
        pointInTimeRecovery: true,
        stream: StreamViewType.NEW_AND_OLD_IMAGES,
        tableName: "blockedCard",
    },
    type: ResourceEnum.DynamoDB,
});

const DYNAMO_CARD_RETRY = STACK.setResource({
    props: {
        partitionKey: {name: "retryKey", type: AttributeType.STRING},
        globalSecondaryIndex: [
            {
                indexName: "cardIdMerchantIndex",
                partitionKey: {name: "cardID", type: AttributeType.STRING},
                sortKey: {name: "merchantID", type: AttributeType.STRING}
            }
        ],
        pointInTimeRecovery: true,
        stream: StreamViewType.NEW_AND_OLD_IMAGES,
        tableName: "cardRetry",
    },
    type: ResourceEnum.DynamoDB,
});

const DEAD_LETTER_BLOCK_CARD_QUEUE: IResourceService<SQSQueueResource> = STACK.setResource<SQSQueueResource>({
    type: ResourceEnum.SQSQueue,
    props: {
        deliveryDelay: Duration.seconds(0),
        queueName: "blockCardDeadLetterQueue",
        visibilityTimeout: Duration.seconds(300), //60
        retentionPeriod: Duration.seconds(1800), //700
    },
})

const BLOCK_CARD_QUEUE: IResourceService<SQSQueueResource> = STACK.setResource<SQSQueueResource>({
    type: ResourceEnum.SQSQueue,
    props: {
        deadLetterQueue: {
            maxReceiveCount: 3,
            queue: DEAD_LETTER_BLOCK_CARD_QUEUE,
        },
        queueName: "blockCardQueue",
        visibilityTimeout: Duration.seconds(300), //180
        retentionPeriod: Duration.seconds(2800) //600
    },
})

const DEAD_LETTER_RESTORE_RETRY_QUEUE: IResourceService<SQSQueueResource> = STACK.setResource<SQSQueueResource>({
    type: ResourceEnum.SQSQueue,
    props: {
        deliveryDelay: Duration.seconds(0),
        queueName: "restoreRetryDeadLetterQueue",
        visibilityTimeout: Duration.seconds(300), //60
        retentionPeriod: Duration.seconds(1800), //700
    },
})

const RESTORE_RETRY_QUEUE: IResourceService<SQSQueueResource> = STACK.setResource<SQSQueueResource>({
    type: ResourceEnum.SQSQueue,
    props: {
        deadLetterQueue: {
            maxReceiveCount: 3,
            queue: DEAD_LETTER_RESTORE_RETRY_QUEUE,
        },
        queueName: "restoreRetryQueue",
        visibilityTimeout: Duration.seconds(300), //180
        retentionPeriod: Duration.seconds(2800) //600
    },
})

// VIRTUAL PRIVATE CLOUD
const VPC_PCI_SUBNETS: IVirtualPrivateCloud =
    STACK.utils.ACCOUNT_ENV === AccountEnvEnum.PROD || STACK.utils.ACCOUNT_ENV === AccountEnvEnum.UAT
        ? {
            securityGroups: [STACK.utils.getEnvDynamodb("VPC_PCI_SG")],
            vpcId: STACK.utils.getEnvDynamodb("VPC_PCI_ID"),
            vpcSubnets: {
                subnets: [
                    STACK.utils.getEnvDynamodb("VPC_PCI_SUBNET_1"),
                    STACK.utils.getEnvDynamodb("VPC_PCI_SUBNET_2"),
                ],
            },
        }
        : {};

const DYNAMO_CARD_INFO = STACK.setResource({
    props: {
        partitionKey: { name: "externalReferenceId", type: AttributeType.STRING },
        globalSecondaryIndex: [
            {
                indexName: "merchantId-index",
                partitionKey: { name: "merchantId", type: AttributeType.STRING },
                sortKey: { name: "createdAt", type: AttributeType.NUMBER }
            },
            {
                indexName: "expiresAt-index",
                partitionKey: { name: "expiresAt", type: AttributeType.NUMBER }
            }
        ],
        pointInTimeRecovery: true,
        stream: StreamViewType.NEW_AND_OLD_IMAGES,
        tableName: "cardInfo",
        timeToLiveAttribute: "expiresAt"  // Automatic cleanup after 180 days
    },
    type: ResourceEnum.DynamoDB,
});

const DEAD_LETTER_CARD_INFO_QUEUE: IResourceService<SQSQueueResource> = STACK.setResource<SQSQueueResource>({
    type: ResourceEnum.SQSQueue,
    props: {
        deliveryDelay: Duration.seconds(0),
        queueName: "cardInfoProcessingDeadLetterQueue",
        visibilityTimeout: Duration.seconds(300),
        retentionPeriod: Duration.seconds(1800), // 30 minutes
    },
})

const CARD_INFO_PROCESSING_QUEUE: IResourceService<SQSQueueResource> = STACK.setResource<SQSQueueResource>({
    type: ResourceEnum.SQSQueue,
    props: {
        deadLetterQueue: {
            maxReceiveCount: 3,
            queue: DEAD_LETTER_CARD_INFO_QUEUE,
        },
        queueName: "cardInfoProcessingQueue",
        visibilityTimeout: Duration.seconds(300), // 5 minutes
        retentionPeriod: Duration.seconds(2800)   // 46+ minutes
    },
})

// Environment
STACK.setEnvironment({
    DYNAMO_BLOCKED_CARD: STACK.utils.getEnvResource(
        DYNAMO_BLOCKED_CARD,
        AttributeTypeEnum.NAME
    ),
    DYNAMO_CARD_RETRY: STACK.utils.getEnvResource(
        DYNAMO_CARD_RETRY,
        AttributeTypeEnum.NAME
    ),
    DYNAMO_CARD_INFO_TABLE: STACK.utils.getEnvResource(
        DYNAMO_CARD_INFO,
        AttributeTypeEnum.NAME
    ),
    ROLLBAR_TOKEN: STACK.utils.getEnvDynamodb("ROLLBAR_TOKEN"),
});

// Plugins
STACK.setPlugins([
    {
        type: PluginsEnum.WARMUP,
        props: {
            schedule: Schedule.rate(cdk.Duration.minutes(5)),
            concurrency: 5,
            payload: "{\"detail\":{\"action\":\"WARMUP\"}}",
        }
    }
])

// Patterns Constants
STACK.setPattern(PatternEnum.SQS_LAMBDA)
    .setEvents([
        {
            type: EventsEnum.QueueEvent,
            props: {
                source: BLOCK_CARD_QUEUE,
                batchSize: 1
            }
        }
    ])
    .setLambda({
        ...LAMBDA_PROPS(
            "blockCard",
            "block_card_handler"
        )
    }).setAccess([
    {
        actions: [DynamoActions.UpdateItem, DynamoActions.GetItem, DynamoActions.PutItem],
        resource: DYNAMO_BLOCKED_CARD
    },
    {
        actions: [DynamoActions.UpdateItem, DynamoActions.GetItem],
        resource: DYNAMO_CARD_RETRY
    }
]);

STACK.setPattern(PatternEnum.SQS_LAMBDA)
    .setEvents([
        {
            type: EventsEnum.QueueEvent,
            props: {
                source: RESTORE_RETRY_QUEUE,
                batchSize: 1
            }
        }
    ])
    .setLambda({
        ...LAMBDA_PROPS(
            "restoreDailyRetries",
            "restore_daily_retries_handler"
        )
    }).setAccess([
    {
        actions: [DynamoActions.UpdateItem, DynamoActions.GetItem],
        resource: DYNAMO_BLOCKED_CARD
    },
    {
        actions: [DynamoActions.Query, DynamoActions.DeleteItem],
        resource: DYNAMO_CARD_RETRY
    }
]);

STACK.setPattern(PatternEnum.SQS_LAMBDA)
    .setEvents([
        {
            type: EventsEnum.QueueEvent,
            props: {
                source: DEAD_LETTER_BLOCK_CARD_QUEUE,
                batchSize: 1
            }
        }
    ])
    .setLambda({
        ...LAMBDA_PROPS(
            "blockCardDLQ",
            "block_card_dlq_handler"
        )
    });

STACK.setPattern(PatternEnum.SQS_LAMBDA)
    .setEvents([
        {
            type: EventsEnum.QueueEvent,
            props: {
                source: DEAD_LETTER_RESTORE_RETRY_QUEUE,
                batchSize: 1
            }
        }
    ])
    .setLambda({
        ...LAMBDA_PROPS(
            "restoreDailyRetriesDLQ",
            "restore_daily_retries_dlq_handler"
        )
    });

STACK.setPattern(PatternEnum.SINGLE_LAMBDA)
    .setLambda({
        ...LAMBDA_PROPS(
            "checkCardStatus",
            "check_card_status_handler"
        ),
        ...VPC_PCI_SUBNETS,
        crossAccount: true
    }).setAccess([
    {
        actions: [DynamoActions.GetItem],
        resource: DYNAMO_BLOCKED_CARD
    },
])

STACK.setPattern(PatternEnum.SQS_LAMBDA)
    .setEvents([
        {
            type: EventsEnum.QueueEvent,
            props: {
                source: CARD_INFO_PROCESSING_QUEUE,
                batchSize: 1  // Process one message at a time for better error handling
            }
        }
    ])
    .setLambda({
        ...LAMBDA_PROPS(
            "cardInfoProcessor",
            "card_info_processor_handler"
        ),
        timeout: Duration.seconds(60), // Longer timeout for encryption operations
        memorySize: 512  // More memory for crypto operations
    })
    .setAccess([
        {
            actions: [DynamoActions.PutItem, DynamoActions.GetItem],
            resource: DYNAMO_CARD_INFO
        }
    ]);

STACK.setPattern(PatternEnum.SQS_LAMBDA)
    .setEvents([
        {
            type: EventsEnum.QueueEvent,
            props: {
                source: DEAD_LETTER_CARD_INFO_QUEUE,
                batchSize: 1
            }
        }
    ])
    .setLambda({
        ...LAMBDA_PROPS(
            "cardInfoProcessorDLQ",
            "card_info_processor_dlq_handler"
        )
    });

// Build
STACK.build();