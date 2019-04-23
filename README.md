# simpletfjob

## 简介

一个简单的模板替换工具,提供一些与tensorflow分布式相关的模板变量,
,用户可以用这些模板变量编写Job模板文件,并且使用该工具将其转换为
Kubernetes可以识别的Job文件。

[下载地址](https://github.com/DQinYuan/simpletfjob/releases/download/0.1/simpletfjob)


## 场景

如果想要使用Tensorflow Estimator API进行分布式训练的话,环境变量一般要求按
如下格式配置

```json
{
  "cluster": {
    "chief": ["10.0.0.4:2222"],
    "ps": ["10.0.0.6:2222"],
    "worker": ["10.0.0.5:2222"]
  },
  "task": {
    "index": 0,
    "type": "ps"
  }
}
```

将容器编排成Job的情况下,每个Job都需要配置不同的环境变量,逐一手动配置
的话相当麻烦,所以这里提供了一些模板变量,工具可以将这些模板变量自动转换
成相应的值。

## 提供的模板的变量

| 变量        | 含义    | 
| --------   | -----   | 
| {{.type}}  | 集群中担任角色,ps或者worker      |  
| {{.index}} | 在集群同类中的编号,从0开始      | 
| {{.host}}  | 物理机的主机名,比如h85     |
| {{.TF_CONFIG}}|  ![little config](https://user-images.githubusercontent.com/23725000/56472553-7dcb7780-6492-11e9-89e3-3c746b91dc6b.png)     |


其中`{{.TF_CONFIG}}`会随着集群中每台机器角色的变化
而发生变化。


比如如下的模板：

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: mnist-{{.type}}-{{.index}}
spec:
  template:
    spec:
      nodeName: {{.host}}
      hostNetwork: true
      containers:
        - name: tensorflow
          image: tensorflow/tensorflow:1.13.1-py3
          command: ["python", "mnist_cnn.py", "--strategy=ps", "--steps=5000"]
          workingDir: /root/share/tensorflow/mnist
          env:
            - name: TF_CONFIG
              value: >
{{.TF_CONFIG}}
          ports:
            - containerPort: 2222
              hostPort: 2222
              name: tfjob-port
          volumeMounts:
            - mountPath: /root/share
              name: mynfs
      restartPolicy: OnFailure
      volumes:
        - name: mynfs
          persistentVolumeClaim:
            claimName: nfs-pvc
  backoffLimit: 4
```

在该模板文件所在目录下执行如下命令：

> 注:该命令一定要在集群的主节点上执行,
> 如果不在主节点上执行的话,切记要先把主节点上的
> `~/.kube/config`文件拷贝到当前机器目录的`~/.kube/`文件夹下

```bash
simpletfjob mnist.yaml -N 3
```

> 注: `-N 3`表示在集群中3台机器上各跑一个tensorflow训练任务进行训练
> 来进行训练。如果不指定`-N`参数
> 的话则默认在集群中每台机器跑一个tensorflow训练任务.

目录下自动生成了一个`mnist_tfjob.yaml`
的文件，这个就是模板处理以后的文件，内容如下：

> 注:程序默认会在原文件名的结尾添加tfjob后缀作为输出文件名


```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: mnist-ps-0
spec:
  template:
    spec:
      nodeName: h10
      hostNetwork: true
      containers:
        - name: tensorflow
          image: tensorflow/tensorflow:1.13.1-py3
          command: ["python", "mnist_cnn.py", "--strategy=ps", "--steps=5000"]
          workingDir: /root/share/tensorflow/mnist
          env:
            - name: TF_CONFIG
              value: >
                {
                  "cluster": {
                    "ps": ["h10:2222"],
                    "worker": ["h100:2222","h14:2222"]
                  },
                  "task": {
                    "index": 0,
                    "type": "ps"
                  }
                }

          ports:
            - containerPort: 2222
              hostPort: 2222
              name: tfjob-port
          volumeMounts:
            - mountPath: /root/share
              name: mynfs
      restartPolicy: OnFailure
      volumes:
        - name: mynfs
          persistentVolumeClaim:
            claimName: nfs-pvc
  backoffLimit: 4
---
apiVersion: batch/v1
kind: Job
metadata:
  name: mnist-worker-0
spec:
  template:
    spec:
      nodeName: h100
      hostNetwork: true
      containers:
        - name: tensorflow
          image: tensorflow/tensorflow:1.13.1-py3
          command: ["python", "mnist_cnn.py", "--strategy=ps", "--steps=5000"]
          workingDir: /root/share/tensorflow/mnist
          env:
            - name: TF_CONFIG
              value: >
                {
                  "cluster": {
                    "ps": ["h10:2222"],
                    "worker": ["h100:2222","h14:2222"]
                  },
                  "task": {
                    "index": 0,
                    "type": "worker"
                  }
                }

          ports:
            - containerPort: 2222
              hostPort: 2222
              name: tfjob-port
          volumeMounts:
            - mountPath: /root/share
              name: mynfs
      restartPolicy: OnFailure
      volumes:
        - name: mynfs
          persistentVolumeClaim:
            claimName: nfs-pvc
  backoffLimit: 4
---
apiVersion: batch/v1
kind: Job
metadata:
  name: mnist-worker-1
spec:
  template:
    spec:
      nodeName: h14
      hostNetwork: true
      containers:
        - name: tensorflow
          image: tensorflow/tensorflow:1.13.1-py3
          command: ["python", "mnist_cnn.py", "--strategy=ps", "--steps=5000"]
          workingDir: /root/share/tensorflow/mnist
          env:
            - name: TF_CONFIG
              value: >
                {
                  "cluster": {
                    "ps": ["h10:2222"],
                    "worker": ["h100:2222","h14:2222"]
                  },
                  "task": {
                    "index": 1,
                    "type": "worker"
                  }
                }

          ports:
            - containerPort: 2222
              hostPort: 2222
              name: tfjob-port
          volumeMounts:
            - mountPath: /root/share
              name: mynfs
      restartPolicy: OnFailure
      volumes:
        - name: mynfs
          persistentVolumeClaim:
            claimName: nfs-pvc
  backoffLimit: 4
```

可以看到生成了3个job，模板变量已经被替换成了相应的值。

如果想排除掉集群中部分正在使用的机器，则只需要在当前目录下新建
一个`exc`文件，里面没行写一个排除掉的主机名，举例如下：

```
h10
h14
h23
h36
h40
```

然后执行命令：

```bash
simpletfjob mnist.yaml -N 3 -E
```

> 注: -E参数会自动读取当前目录下的`exc`文件并在生成
> 文件时将这些主机排除掉

## 其他参数

| 全称        | 简称    | 含义 | 默认值|
| --------   | -----   | ----- | ----- |
| --help | -h| 查看帮助信息| 无|
| --exc  | -E      | 文件名,该文件中记录的主机都会被排除在训练机之外  | 无 |
| --num | -N      | 训练使用的服务器数目 | -1(表示使用全部服务器) |
| --psn  | 无     | 训练使用的parameter server的数目 | 1 |
| --psf|  无    | 暂时没有实现该参数,请不要使用 | 无 |

