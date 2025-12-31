// Package agentcore provides Pulumi components for AgentCore deployments on AWS.
package agentcore

import (
	"fmt"

	"github.com/agentplexus/agentkit/platforms/agentcore/iac"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// mergeTags merges base tags with additional name tag.
func mergeTags(base pulumi.StringMap, name pulumi.StringInput) pulumi.StringMap {
	result := pulumi.StringMap{}
	for k, v := range base {
		result[k] = v
	}
	result["Name"] = name
	return result
}

// AgentCoreStack contains all the Pulumi resources for an AgentCore deployment.
type AgentCoreStack struct {
	// Config is the stack configuration.
	Config iac.StackConfig

	// VPC is the VPC resource (nil if using existing VPC).
	VPC *ec2.Vpc

	// PublicSubnet is the public subnet.
	PublicSubnet *ec2.Subnet

	// PrivateSubnet is the private subnet.
	PrivateSubnet *ec2.Subnet

	// InternetGateway is the internet gateway.
	InternetGateway *ec2.InternetGateway

	// NatGateway is the NAT gateway.
	NatGateway *ec2.NatGateway

	// SecurityGroup is the security group for agents.
	SecurityGroup *ec2.SecurityGroup

	// ExecutionRole is the IAM execution role.
	ExecutionRole *iam.Role

	// LogGroup is the CloudWatch log group.
	LogGroup *cloudwatch.LogGroup

	// Outputs contains stack output values.
	Outputs map[string]pulumi.StringOutput
}

// NewAgentCoreStack creates all AgentCore resources from a StackConfig.
func NewAgentCoreStack(ctx *pulumi.Context, config iac.StackConfig) (*AgentCoreStack, error) {
	// Validate and apply defaults
	config.ApplyDefaults()
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid stack configuration: %w", err)
	}

	stack := &AgentCoreStack{
		Config:  config,
		Outputs: make(map[string]pulumi.StringOutput),
	}

	// Create tags map
	tags := pulumi.StringMap{}
	for k, v := range config.Tags {
		tags[k] = pulumi.String(v)
	}
	tags["ManagedBy"] = pulumi.String("agentkit-pulumi")

	// Create VPC resources
	if config.VPC.CreateVPC {
		if err := stack.createVPC(ctx, tags); err != nil {
			return nil, fmt.Errorf("failed to create VPC: %w", err)
		}
	}

	// Create security group
	if err := stack.createSecurityGroup(ctx, tags); err != nil {
		return nil, fmt.Errorf("failed to create security group: %w", err)
	}

	// Create IAM role
	if err := stack.createIAMRole(ctx, tags); err != nil {
		return nil, fmt.Errorf("failed to create IAM role: %w", err)
	}

	// Create CloudWatch log group
	if config.Observability.EnableCloudWatchLogs {
		if err := stack.createLogGroup(ctx, tags); err != nil {
			return nil, fmt.Errorf("failed to create log group: %w", err)
		}
	}

	// Export outputs
	stack.exportOutputs(ctx)

	return stack, nil
}

// createVPC creates VPC and networking resources.
func (s *AgentCoreStack) createVPC(ctx *pulumi.Context, tags pulumi.StringMap) error {
	var err error
	stackName := s.Config.StackName

	// Create VPC
	s.VPC, err = ec2.NewVpc(ctx, "vpc", &ec2.VpcArgs{
		CidrBlock:          pulumi.String(s.Config.VPC.VPCCidr),
		EnableDnsHostnames: pulumi.Bool(true),
		EnableDnsSupport:   pulumi.Bool(true),
		Tags:               mergeTags(tags, pulumi.Sprintf("%s-vpc", stackName)),
	})
	if err != nil {
		return err
	}

	// Create Internet Gateway
	s.InternetGateway, err = ec2.NewInternetGateway(ctx, "igw", &ec2.InternetGatewayArgs{
		VpcId: s.VPC.ID(),
		Tags:  mergeTags(tags, pulumi.Sprintf("%s-igw", stackName)),
	})
	if err != nil {
		return err
	}

	// Create public subnet
	s.PublicSubnet, err = ec2.NewSubnet(ctx, "public-subnet", &ec2.SubnetArgs{
		VpcId:               s.VPC.ID(),
		CidrBlock:           pulumi.String("10.0.1.0/24"),
		MapPublicIpOnLaunch: pulumi.Bool(true),
		Tags:                mergeTags(tags, pulumi.Sprintf("%s-public", stackName)),
	})
	if err != nil {
		return err
	}

	// Create private subnet
	s.PrivateSubnet, err = ec2.NewSubnet(ctx, "private-subnet", &ec2.SubnetArgs{
		VpcId:     s.VPC.ID(),
		CidrBlock: pulumi.String("10.0.10.0/24"),
		Tags:      mergeTags(tags, pulumi.Sprintf("%s-private", stackName)),
	})
	if err != nil {
		return err
	}

	// Create Elastic IP for NAT Gateway
	eip, err := ec2.NewEip(ctx, "nat-eip", &ec2.EipArgs{
		Domain: pulumi.String("vpc"),
		Tags:   mergeTags(tags, pulumi.Sprintf("%s-nat-eip", stackName)),
	}, pulumi.DependsOn([]pulumi.Resource{s.InternetGateway}))
	if err != nil {
		return err
	}

	// Create NAT Gateway
	s.NatGateway, err = ec2.NewNatGateway(ctx, "nat", &ec2.NatGatewayArgs{
		AllocationId: eip.ID(),
		SubnetId:     s.PublicSubnet.ID(),
		Tags:         mergeTags(tags, pulumi.Sprintf("%s-nat", stackName)),
	}, pulumi.DependsOn([]pulumi.Resource{s.InternetGateway}))
	if err != nil {
		return err
	}

	// Create public route table
	publicRouteTable, err := ec2.NewRouteTable(ctx, "public-rt", &ec2.RouteTableArgs{
		VpcId: s.VPC.ID(),
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String("0.0.0.0/0"),
				GatewayId: s.InternetGateway.ID(),
			},
		},
		Tags: mergeTags(tags, pulumi.Sprintf("%s-public-rt", stackName)),
	})
	if err != nil {
		return err
	}

	// Associate public subnet with public route table
	_, err = ec2.NewRouteTableAssociation(ctx, "public-rta", &ec2.RouteTableAssociationArgs{
		SubnetId:     s.PublicSubnet.ID(),
		RouteTableId: publicRouteTable.ID(),
	})
	if err != nil {
		return err
	}

	// Create private route table
	privateRouteTable, err := ec2.NewRouteTable(ctx, "private-rt", &ec2.RouteTableArgs{
		VpcId: s.VPC.ID(),
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock:    pulumi.String("0.0.0.0/0"),
				NatGatewayId: s.NatGateway.ID(),
			},
		},
		Tags: mergeTags(tags, pulumi.Sprintf("%s-private-rt", stackName)),
	})
	if err != nil {
		return err
	}

	// Associate private subnet with private route table
	_, err = ec2.NewRouteTableAssociation(ctx, "private-rta", &ec2.RouteTableAssociationArgs{
		SubnetId:     s.PrivateSubnet.ID(),
		RouteTableId: privateRouteTable.ID(),
	})
	if err != nil {
		return err
	}

	return nil
}

// createSecurityGroup creates the security group for agents.
func (s *AgentCoreStack) createSecurityGroup(ctx *pulumi.Context, tags pulumi.StringMap) error {
	var err error
	stackName := s.Config.StackName

	var vpcId pulumi.StringInput
	if s.VPC != nil {
		vpcId = s.VPC.ID()
	} else if s.Config.VPC.VPCID != "" {
		vpcId = pulumi.String(s.Config.VPC.VPCID)
	}

	s.SecurityGroup, err = ec2.NewSecurityGroup(ctx, "sg", &ec2.SecurityGroupArgs{
		Name:        pulumi.Sprintf("%s-sg", stackName),
		Description: pulumi.Sprintf("Security group for %s AgentCore agents", stackName),
		VpcId:       vpcId,
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: mergeTags(tags, pulumi.Sprintf("%s-sg", stackName)),
	})
	if err != nil {
		return err
	}

	// Add self-referencing ingress rule for agent-to-agent communication
	_, err = ec2.NewSecurityGroupRule(ctx, "sg-self-ingress", &ec2.SecurityGroupRuleArgs{
		Type:                  pulumi.String("ingress"),
		SecurityGroupId:       s.SecurityGroup.ID(),
		SourceSecurityGroupId: s.SecurityGroup.ID(),
		Protocol:              pulumi.String("-1"),
		FromPort:              pulumi.Int(0),
		ToPort:                pulumi.Int(0),
		Description:           pulumi.String("Allow communication between agents"),
	})
	if err != nil {
		return err
	}

	return nil
}

// createIAMRole creates the IAM execution role for agents.
func (s *AgentCoreStack) createIAMRole(ctx *pulumi.Context, tags pulumi.StringMap) error {
	var err error
	stackName := s.Config.StackName

	// Create assume role policy
	assumeRolePolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"Service": ["bedrock.amazonaws.com", "lambda.amazonaws.com"]
				},
				"Action": "sts:AssumeRole"
			}
		]
	}`

	s.ExecutionRole, err = iam.NewRole(ctx, "execution-role", &iam.RoleArgs{
		Name:             pulumi.Sprintf("%s-execution-role", stackName),
		Description:      pulumi.Sprintf("Execution role for %s AgentCore agents", stackName),
		AssumeRolePolicy: pulumi.String(assumeRolePolicy),
		Tags:             mergeTags(tags, pulumi.Sprintf("%s-execution-role", stackName)),
	})
	if err != nil {
		return err
	}

	// Build IAM policy statements
	policyStatements := s.buildIAMPolicyStatements()

	// Create and attach policy
	policy, err := iam.NewPolicy(ctx, "execution-policy", &iam.PolicyArgs{
		Name:        pulumi.Sprintf("%s-execution-policy", stackName),
		Description: pulumi.Sprintf("Execution policy for %s AgentCore agents", stackName),
		Policy:      pulumi.String(policyStatements),
	})
	if err != nil {
		return err
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "execution-policy-attachment", &iam.RolePolicyAttachmentArgs{
		Role:      s.ExecutionRole.Name,
		PolicyArn: policy.Arn,
	})
	if err != nil {
		return err
	}

	return nil
}

// buildIAMPolicyStatements builds the IAM policy JSON.
func (s *AgentCoreStack) buildIAMPolicyStatements() string {
	statements := []string{
		// CloudWatch Logs
		`{
			"Effect": "Allow",
			"Action": [
				"logs:CreateLogGroup",
				"logs:CreateLogStream",
				"logs:PutLogEvents"
			],
			"Resource": "arn:aws:logs:*:*:*"
		}`,
		// ECR
		`{
			"Effect": "Allow",
			"Action": [
				"ecr:GetAuthorizationToken",
				"ecr:BatchCheckLayerAvailability",
				"ecr:GetDownloadUrlForLayer",
				"ecr:BatchGetImage"
			],
			"Resource": "*"
		}`,
	}

	// Bedrock access
	if s.Config.IAM.EnableBedrockAccess {
		bedrockResource := `"arn:aws:bedrock:*:*:foundation-model/*"`
		if len(s.Config.IAM.BedrockModelIDs) > 0 {
			resources := ""
			for i, modelID := range s.Config.IAM.BedrockModelIDs {
				if i > 0 {
					resources += ", "
				}
				resources += fmt.Sprintf(`"arn:aws:bedrock:*:*:foundation-model/%s"`, modelID)
			}
			bedrockResource = fmt.Sprintf("[%s]", resources)
		}
		statements = append(statements, fmt.Sprintf(`{
			"Effect": "Allow",
			"Action": [
				"bedrock:InvokeModel",
				"bedrock:InvokeModelWithResponseStream"
			],
			"Resource": %s
		}`, bedrockResource))
	}

	// Secrets Manager access
	hasSecrets := false
	for _, agent := range s.Config.Agents {
		if len(agent.SecretsARNs) > 0 {
			hasSecrets = true
			break
		}
	}
	if hasSecrets {
		statements = append(statements, `{
			"Effect": "Allow",
			"Action": [
				"secretsmanager:GetSecretValue"
			],
			"Resource": "*"
		}`)
	}

	// Build final policy
	statementsJSON := ""
	for i, stmt := range statements {
		if i > 0 {
			statementsJSON += ","
		}
		statementsJSON += stmt
	}

	return fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [%s]
	}`, statementsJSON)
}

// createLogGroup creates the CloudWatch log group.
func (s *AgentCoreStack) createLogGroup(ctx *pulumi.Context, tags pulumi.StringMap) error {
	var err error
	stackName := s.Config.StackName

	retentionDays := s.Config.Observability.LogRetentionDays
	if retentionDays == 0 {
		retentionDays = 30
	}

	s.LogGroup, err = cloudwatch.NewLogGroup(ctx, "log-group", &cloudwatch.LogGroupArgs{
		Name:            pulumi.Sprintf("/aws/agentcore/%s", stackName),
		RetentionInDays: pulumi.Int(retentionDays),
		Tags:            mergeTags(tags, pulumi.Sprintf("%s-logs", stackName)),
	})
	if err != nil {
		return err
	}

	return nil
}

// exportOutputs exports stack outputs.
func (s *AgentCoreStack) exportOutputs(ctx *pulumi.Context) {
	if s.VPC != nil {
		ctx.Export("vpcId", s.VPC.ID())
		s.Outputs["vpcId"] = s.VPC.ID().ToStringOutput()
	}

	if s.PrivateSubnet != nil {
		ctx.Export("privateSubnetId", s.PrivateSubnet.ID())
		s.Outputs["privateSubnetId"] = s.PrivateSubnet.ID().ToStringOutput()
	}

	if s.SecurityGroup != nil {
		ctx.Export("securityGroupId", s.SecurityGroup.ID())
		s.Outputs["securityGroupId"] = s.SecurityGroup.ID().ToStringOutput()
	}

	if s.ExecutionRole != nil {
		ctx.Export("executionRoleArn", s.ExecutionRole.Arn)
		s.Outputs["executionRoleArn"] = s.ExecutionRole.Arn
	}

	if s.LogGroup != nil {
		ctx.Export("logGroupName", s.LogGroup.Name)
		s.Outputs["logGroupName"] = s.LogGroup.Name
	}

	ctx.Export("agentCount", pulumi.Int(len(s.Config.Agents)))
}

// NewStackFromFile creates an AgentCoreStack from a JSON or YAML config file.
func NewStackFromFile(ctx *pulumi.Context, configPath string) (*AgentCoreStack, error) {
	config, err := iac.LoadStackConfigFromFile(configPath)
	if err != nil {
		return nil, err
	}
	return NewAgentCoreStack(ctx, *config)
}

// MustNewStackFromFile is like NewStackFromFile but panics on error.
func MustNewStackFromFile(ctx *pulumi.Context, configPath string) *AgentCoreStack {
	stack, err := NewStackFromFile(ctx, configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to create stack from %s: %v", configPath, err))
	}
	return stack
}
