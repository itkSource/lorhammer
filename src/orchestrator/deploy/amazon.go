package deploy

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"lorhammer/src/tools"
)

const TypeAmazon = "amazon"

var _LOG_AMAZON = logrus.WithField("logger", "orchestrator/deploy/amazon")

type amazonImpl struct {
	Region           string          `json:"region"`
	ImageId          string          `json:"imageId"`
	InstanceType     string          `json:"instanceType"`
	KeyPairName      string          `json:"keyPairName"`
	SecurityGroupIds []string        `json:"securityGroupIds"`
	MinCount         int64           `json:"minCount"`
	MaxCount         int64           `json:"maxCount"`
	DistantConfig    json.RawMessage `json:"distantConfig"`

	ec2Client       *ec2.EC2
	distantDeployer *distantImpl
	consulAddress   string
	instancesId     []*string
}

func NewAmazonFromJson(serialized json.RawMessage, consulClient tools.Consul) (Deployer, error) {
	client := &amazonImpl{}
	if err := json.Unmarshal(serialized, client); err != nil {
		return nil, err
	}

	b := true
	s, err := session.NewSession(&aws.Config{Region: aws.String(client.Region), CredentialsChainVerboseErrors: &b})
	if err != nil {
		return nil, err
	}

	client.ec2Client = ec2.New(s)
	if client.distantDeployer, err = newDistantImpl(client.DistantConfig); err != nil {
		return nil, err
	}
	client.consulAddress = consulClient.GetAddress()
	return client, nil
}

func (client *amazonImpl) RunBefore() error {
	runResult, err := client.ec2Client.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String(client.ImageId),
		InstanceType:     aws.String(client.InstanceType),
		KeyName:          aws.String(client.KeyPairName),
		SecurityGroupIds: aws.StringSlice(client.SecurityGroupIds),
		MinCount:         aws.Int64(client.MinCount),
		MaxCount:         aws.Int64(client.MaxCount),
	})
	if err != nil {
		return err
	}

	instancesId := make([]*string, len(runResult.Instances))
	for index, instance := range runResult.Instances {
		instancesId[index] = instance.InstanceId
	}
	client.instancesId = instancesId

	_LOG_AMAZON.WithField("nb", len(runResult.Instances)).Info("Created instance on amazon wait until running")
	if err := client.ec2Client.WaitUntilInstanceStatusOk(&ec2.DescribeInstanceStatusInput{InstanceIds: instancesId}); err != nil {
		return err
	}

	_LOG_AMAZON.Info("Configure networks")
	for _, instance := range runResult.Instances {
		for _, network := range instance.NetworkInterfaces {
			client.ec2Client.ModifyNetworkInterfaceAttribute(&ec2.ModifyNetworkInterfaceAttributeInput{
				NetworkInterfaceId: network.NetworkInterfaceId,
				SourceDestCheck:    &ec2.AttributeBooleanValue{Value: aws.Bool(false)},
			})
		}
	}
	_LOG_AMAZON.WithField("nb", len(runResult.Instances)).Info("Created instance on amazon running")

	return nil
}

func (client *amazonImpl) Deploy() error {
	return nil // deploy and launch are grouped in RunAfter to not make multiple call on distant server
}

func (client *amazonImpl) RunAfter() error {
	res, err := client.ec2Client.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: client.instancesId})
	if err != nil {
		return err
	}

	_LOG_AMAZON.WithField("nb", len(client.instancesId)).Info("Deploy lorhammer")

	for _, reservation := range res.Reservations {
		for _, instance := range reservation.Instances {
			client.distantDeployer.IpServer = *instance.PublicDnsName
			client.distantDeployer.AfterCmd = fmt.Sprintf("nohup %s/lorhammer -consul %s -local-ip %s > lorahmmer.log 2>&1 &", client.distantDeployer.PathWhereScp, client.consulAddress, *instance.PublicDnsName)
			err := client.distantDeployer.Deploy()
			if err != nil {
				_LOG_AMAZON.WithError(err).Error("Lorhammer not deployed")
			} else {
				_LOG_AMAZON.Info("Lorhammer deployed")
			}

			err = client.distantDeployer.RunAfter()
			if err != nil {
				_LOG_AMAZON.WithError(err).Error("Lorhammer not started")
			} else {
				_LOG_AMAZON.Info("Lorhammers started")
			}
		}
	}

	_LOG_AMAZON.WithField("nb", len(client.instancesId)).Info("All Lorhammers deployed !!")

	return nil
}
